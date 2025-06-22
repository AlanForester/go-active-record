package activerecord

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxOpen  int
	MaxIdle  int
	Lifetime time.Duration
}

// DatabaseConnection represents a database connection
type DatabaseConnection struct {
	config *DatabaseConfig
	db     *sql.DB
	mu     sync.RWMutex
}

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection(config *DatabaseConfig) (*DatabaseConnection, error) {
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxOpen)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(config.Lifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseConnection{
		config: config,
		db:     db,
	}, nil
}

// GetDB returns the underlying database connection
func (dc *DatabaseConnection) GetDB() *sql.DB {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.db
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if dc.db != nil {
		return dc.db.Close()
	}
	return nil
}

// HealthCheck performs a health check on the database
func (dc *DatabaseConnection) HealthCheck() error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.db.Ping()
}

// DatabaseType represents the type of database operation
type DatabaseType int

const (
	Primary DatabaseType = iota
	ReadReplica
	WriteReplica
)

// DatabaseResolver manages multiple database connections
type DatabaseResolver struct {
	primary       *DatabaseConnection
	readReplicas  []*DatabaseConnection
	writeReplicas []*DatabaseConnection
	mu            sync.RWMutex
	roundRobin    int
}

// NewDatabaseResolver creates a new database resolver
func NewDatabaseResolver() *DatabaseResolver {
	return &DatabaseResolver{
		readReplicas:  make([]*DatabaseConnection, 0),
		writeReplicas: make([]*DatabaseConnection, 0),
		roundRobin:    0,
	}
}

// SetPrimary sets the primary database connection
func (dr *DatabaseResolver) SetPrimary(config *DatabaseConfig) error {
	conn, err := NewDatabaseConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create primary connection: %w", err)
	}

	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.primary = conn
	return nil
}

// AddReadReplica adds a read replica
func (dr *DatabaseResolver) AddReadReplica(config *DatabaseConfig) error {
	conn, err := NewDatabaseConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create read replica connection: %w", err)
	}

	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.readReplicas = append(dr.readReplicas, conn)
	return nil
}

// AddWriteReplica adds a write replica
func (dr *DatabaseResolver) AddWriteReplica(config *DatabaseConfig) error {
	conn, err := NewDatabaseConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create write replica connection: %w", err)
	}

	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.writeReplicas = append(dr.writeReplicas, conn)
	return nil
}

// GetConnection returns the appropriate database connection based on operation type
func (dr *DatabaseResolver) GetConnection(dbType DatabaseType) (*sql.DB, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	switch dbType {
	case Primary:
		if dr.primary == nil {
			return nil, fmt.Errorf("primary database not configured")
		}
		return dr.primary.GetDB(), nil
	case ReadReplica:
		if len(dr.readReplicas) == 0 {
			// Fallback to primary if no read replicas
			if dr.primary == nil {
				return nil, fmt.Errorf("no read replicas configured and primary not available")
			}
			return dr.primary.GetDB(), nil
		}
		// Round-robin selection
		conn := dr.readReplicas[dr.roundRobin%len(dr.readReplicas)]
		dr.roundRobin++
		return conn.GetDB(), nil
	case WriteReplica:
		if len(dr.writeReplicas) == 0 {
			// Fallback to primary if no write replicas
			if dr.primary == nil {
				return nil, fmt.Errorf("no write replicas configured and primary not available")
			}
			return dr.primary.GetDB(), nil
		}
		// Round-robin selection
		conn := dr.writeReplicas[dr.roundRobin%len(dr.writeReplicas)]
		dr.roundRobin++
		return conn.GetDB(), nil
	default:
		return nil, fmt.Errorf("unknown database type")
	}
}

// GetReadConnection returns a read connection (read replica or primary)
func (dr *DatabaseResolver) GetReadConnection() (*sql.DB, error) {
	return dr.GetConnection(ReadReplica)
}

// GetWriteConnection returns a write connection (write replica or primary)
func (dr *DatabaseResolver) GetWriteConnection() (*sql.DB, error) {
	return dr.GetConnection(WriteReplica)
}

// GetPrimaryConnection returns the primary connection
func (dr *DatabaseResolver) GetPrimaryConnection() (*sql.DB, error) {
	return dr.GetConnection(Primary)
}

// HealthCheck performs health checks on all connections
func (dr *DatabaseResolver) HealthCheck() map[string]error {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	results := make(map[string]error)

	// Check primary
	if dr.primary != nil {
		results["primary"] = dr.primary.HealthCheck()
	}

	// Check read replicas
	for i, replica := range dr.readReplicas {
		results[fmt.Sprintf("read_replica_%d", i)] = replica.HealthCheck()
	}

	// Check write replicas
	for i, replica := range dr.writeReplicas {
		results[fmt.Sprintf("write_replica_%d", i)] = replica.HealthCheck()
	}

	return results
}

// Close closes all database connections
func (dr *DatabaseResolver) Close() error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	var errors []error

	// Close primary
	if dr.primary != nil {
		if err := dr.primary.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close primary: %w", err))
		}
	}

	// Close read replicas
	for i, replica := range dr.readReplicas {
		if err := replica.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close read replica %d: %w", i, err))
		}
	}

	// Close write replicas
	for i, replica := range dr.writeReplicas {
		if err := replica.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close write replica %d: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}

// DatabaseManager manages multiple databases
type DatabaseManager struct {
	resolvers map[string]*DatabaseResolver
	mu        sync.RWMutex
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		resolvers: make(map[string]*DatabaseResolver),
	}
}

// AddDatabase adds a database with a name
func (dm *DatabaseManager) AddDatabase(name string, resolver *DatabaseResolver) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.resolvers[name] = resolver
}

// GetDatabase returns a database resolver by name
func (dm *DatabaseManager) GetDatabase(name string) (*DatabaseResolver, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	resolver, exists := dm.resolvers[name]
	if !exists {
		return nil, fmt.Errorf("database '%s' not found", name)
	}

	return resolver, nil
}

// GetConnection returns a connection from a specific database
func (dm *DatabaseManager) GetConnection(databaseName string, dbType DatabaseType) (*sql.DB, error) {
	resolver, err := dm.GetDatabase(databaseName)
	if err != nil {
		return nil, err
	}

	return resolver.GetConnection(dbType)
}

// HealthCheck performs health checks on all databases
func (dm *DatabaseManager) HealthCheck() map[string]map[string]error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	results := make(map[string]map[string]error)

	for name, resolver := range dm.resolvers {
		results[name] = resolver.HealthCheck()
	}

	return results
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var errors []error

	for name, resolver := range dm.resolvers {
		if err := resolver.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database '%s': %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// Global database manager instance
var globalDatabaseManager *DatabaseManager

// SetDatabaseManager sets the global database manager
func SetDatabaseManager(dm *DatabaseManager) {
	globalDatabaseManager = dm
}

// GetDatabaseManager returns the global database manager
func GetDatabaseManager() *DatabaseManager {
	if globalDatabaseManager == nil {
		globalDatabaseManager = NewDatabaseManager()
	}
	return globalDatabaseManager
}

// Database-aware query execution functions

// ExecOnDatabase executes a query on a specific database
func ExecOnDatabase(databaseName string, dbType DatabaseType, query string, args ...interface{}) (sql.Result, error) {
	db, err := GetDatabaseManager().GetConnection(databaseName, dbType)
	if err != nil {
		return nil, err
	}

	return db.Exec(query, args...)
}

// QueryOnDatabase executes a query on a specific database
func QueryOnDatabase(databaseName string, dbType DatabaseType, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := GetDatabaseManager().GetConnection(databaseName, dbType)
	if err != nil {
		return nil, err
	}

	return db.Query(query, args...)
}

// QueryRowOnDatabase executes a query on a specific database and returns a single row
func QueryRowOnDatabase(databaseName string, dbType DatabaseType, query string, args ...interface{}) *sql.Row {
	db, err := GetDatabaseManager().GetConnection(databaseName, dbType)
	if err != nil {
		return nil
	}

	return db.QueryRow(query, args...)
}

// BeginTransactionOnDatabase begins a transaction on a specific database
func BeginTransactionOnDatabase(databaseName string, dbType DatabaseType) (*sql.Tx, error) {
	db, err := GetDatabaseManager().GetConnection(databaseName, dbType)
	if err != nil {
		return nil, err
	}

	return db.Begin()
}

// BeginTransactionOnDatabaseWithContext begins a transaction with context on a specific database
func BeginTransactionOnDatabaseWithContext(ctx context.Context, databaseName string, dbType DatabaseType) (*sql.Tx, error) {
	db, err := GetDatabaseManager().GetConnection(databaseName, dbType)
	if err != nil {
		return nil, err
	}

	return db.BeginTx(ctx, nil)
}

// Database-aware model operations

// CreateOnDatabase creates a record on a specific database
func CreateOnDatabase(databaseName string, model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	// Set timestamps
	now := time.Now()
	modeler.SetCreatedAt(now)
	modeler.SetUpdatedAt(now)

	// Get fields and values
	fields, values := getFieldsAndValues(model, false)
	if len(fields) == 0 {
		return fmt.Errorf("no fields to insert")
	}

	// Build query
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		modeler.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute query on write database
	result, err := ExecOnDatabase(databaseName, WriteReplica, query, values...)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	// Set the generated ID
	if id, err := result.LastInsertId(); err == nil {
		modeler.SetID(id)
	}

	return nil
}

// FindOnDatabase finds a record on a specific database
func FindOnDatabase(databaseName string, model interface{}, id interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", modeler.TableName())
	rows, err := QueryOnDatabase(databaseName, ReadReplica, query, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}
	return scanRow(rows, model)
}

// UpdateOnDatabase updates a record on a specific database
func UpdateOnDatabase(databaseName string, model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	// Set updated timestamp
	modeler.SetUpdatedAt(time.Now())

	// Get fields and values
	fields, values := getFieldsAndValues(model, true)
	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Build query
	setClause := make([]string, len(fields))
	for i, field := range fields {
		setClause[i] = fmt.Sprintf("%s = ?", field)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		modeler.TableName(),
		strings.Join(setClause, ", "),
	)

	// Add ID to values
	values = append(values, modeler.GetID())

	// Execute query on write database
	_, err := ExecOnDatabase(databaseName, WriteReplica, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// DeleteOnDatabase deletes a record on a specific database
func DeleteOnDatabase(databaseName string, model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", modeler.TableName())
	_, err := ExecOnDatabase(databaseName, WriteReplica, query, modeler.GetID())
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}
