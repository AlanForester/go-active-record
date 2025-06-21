package activerecord

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Migration represents a database migration.
type Migration struct {
	ID        int       `db:"id"`
	Version   string    `db:"version"`
	CreatedAt time.Time `db:"created_at"`
}

// Migrator interface for migrations.
type Migrator interface {
	Up() error
	Down() error
	Version() string
}

// MigrationManager manages database migrations.
type MigrationManager struct {
	migrations []Migrator
}

// NewMigrationManager creates a new migration manager.
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: make([]Migrator, 0),
	}
}

// AddMigration adds a migration to the manager.
func (mm *MigrationManager) AddMigration(migration Migrator) {
	mm.migrations = append(mm.migrations, migration)
}

// Migrate runs all pending migrations.
func (mm *MigrationManager) Migrate() error {
	// Create migrations table if it doesn't exist.
	if err := mm.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations.
	applied, err := mm.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations.
	for _, migration := range mm.migrations {
		version := migration.Version()
		if !applied[version] {
			if err := migration.Up(); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", version, err)
			}
			if err := mm.recordMigration(version); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", version, err)
			}
		}
	}

	return nil
}

// Rollback rolls back the last migration.
func (mm *MigrationManager) Rollback() error {
	// Get the last applied migration.
	lastMigration, err := mm.getLastMigration()
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the migration and roll it back.
	for _, migration := range mm.migrations {
		if migration.Version() == lastMigration.Version {
			if err := migration.Down(); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Version, err)
			}
			if err := mm.removeMigration(lastMigration.Version); err != nil {
				return fmt.Errorf("failed to remove migration record %s: %w", lastMigration.Version, err)
			}
			return nil
		}
	}

	return fmt.Errorf("migration %s not found", lastMigration.Version)
}

// createMigrationsTable creates the migrations table.
func (mm *MigrationManager) createMigrationsTable() error {
	query := `CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions.
func (mm *MigrationManager) getAppliedMigrations() (map[string]bool, error) {
	rows, err := Query("SELECT version FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}

// getLastMigration returns the last applied migration.
func (mm *MigrationManager) getLastMigration() (*Migration, error) {
	row := QueryRow("SELECT id, version, created_at FROM migrations ORDER BY id DESC LIMIT 1")
	if row == nil {
		return nil, nil
	}

	var migration Migration
	err := row.Scan(&migration.ID, &migration.Version, &migration.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &migration, nil
}

// recordMigration records a migration as applied.
func (mm *MigrationManager) recordMigration(version string) error {
	_, err := Exec("INSERT INTO migrations (version) VALUES (?)", version)
	return err
}

// removeMigration removes a migration record.
func (mm *MigrationManager) removeMigration(version string) error {
	_, err := Exec("DELETE FROM migrations WHERE version = ?", version)
	return err
}

// Migration interface for migrations
type MigrationInterface interface {
	Up() error
	Down() error
	Version() int64
}

// MigrationRecord record of a completed migration
type MigrationRecord struct {
	Version   int64     `db:"version"`
	AppliedAt time.Time `db:"applied_at"`
}

// SchemaMigration manages schema migrations
type SchemaMigration struct {
	ActiveRecordModel
}

// TableName returns the name of the migrations table
func (sm *SchemaMigration) TableName() string {
	return "schema_migrations"
}

// Migrator manages migrations
type MigratorStruct struct {
	db *sql.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator() *MigratorStruct {
	return &MigratorStruct{db: GetConnection()}
}

// CreateMigrationsTable creates a table for tracking migrations
func (m *MigratorStruct) CreateMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.Exec(query)
	return err
}

// Migrate performs all unapplied migrations
func (m *MigratorStruct) Migrate(migrations []MigrationInterface) error {
	// Create migrations table if it doesn't exist
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Perform migrations
	for _, migration := range migrations {
		if !m.isMigrationApplied(applied, migration.Version()) {
			if err := m.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version(), err)
			}
		}
	}

	return nil
}

// Rollback rolls back the last migration
func (m *MigratorStruct) Rollback(migrations []MigrationInterface) error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no applied migrations to rollback")
	}

	// Find the last migration
	lastMigration := applied[len(applied)-1]

	// Find the corresponding migration in the list
	var targetMigration MigrationInterface
	for _, migration := range migrations {
		if migration.Version() == lastMigration.Version {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found", lastMigration.Version)
	}

	// Perform rollback
	if err := targetMigration.Down(); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", lastMigration.Version, err)
	}

	// Remove migration record
	if err := m.removeMigrationRecord(lastMigration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

// Status shows the status of migrations
func (m *MigratorStruct) Status(migrations []MigrationInterface) error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("==================")

	for _, migration := range migrations {
		status := "down"
		if m.isMigrationApplied(applied, migration.Version()) {
			status = "up"
		}
		fmt.Printf("%d\t%s\n", migration.Version(), status)
	}

	return nil
}

// Helper methods

func (m *MigratorStruct) getAppliedMigrations() ([]MigrationRecord, error) {
	query := "SELECT version, applied_at FROM schema_migrations ORDER BY version"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, record)
	}

	return migrations, rows.Err()
}

func (m *MigratorStruct) isMigrationApplied(applied []MigrationRecord, version int64) bool {
	for _, record := range applied {
		if record.Version == version {
			return true
		}
	}
	return false
}

func (m *MigratorStruct) runMigration(migration MigrationInterface) error {
	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			// Log the rollback error but don't return it
			// as it would mask the original error
			_ = err // Intentionally ignoring rollback error
		}
	}()

	// Perform migration
	if err := migration.Up(); err != nil {
		return err
	}

	// Add migration record
	query := "INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)"
	_, err = tx.Exec(query, migration.Version(), time.Now())
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

func (m *MigratorStruct) removeMigrationRecord(version int64) error {
	query := "DELETE FROM schema_migrations WHERE version = ?"
	_, err := m.db.Exec(query, version)
	return err
}

// Schema Builder methods

// CreateTable creates a table
func CreateTable(tableName string, callback func(*TableBuilder)) error {
	builder := &TableBuilder{tableName: tableName}
	callback(builder)

	query, indexes := builder.Build()
	_, err := GetConnection().Exec(query)
	if err != nil {
		return err
	}
	for _, idx := range indexes {
		_, err := GetConnection().Exec(idx)
		if err != nil {
			return err
		}
	}
	return nil
}

// DropTable deletes a table
func DropTable(tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := GetConnection().Exec(query)
	return err
}

// TableBuilder table builder
type TableBuilder struct {
	tableName string
	columns   []string
	indexes   []string
}

// Column adds a column
func (tb *TableBuilder) Column(name, dataType string, options ...string) {
	column := fmt.Sprintf("%s %s", name, dataType)
	if len(options) > 0 {
		column += " " + strings.Join(options, " ")
	}
	tb.columns = append(tb.columns, column)
}

// PrimaryKey adds a primary key
func (tb *TableBuilder) PrimaryKey(columns ...string) {
	key := fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(columns, ", "))
	tb.columns = append(tb.columns, key)
}

// Index adds an index
func (tb *TableBuilder) Index(columns ...string) {
	indexName := fmt.Sprintf("idx_%s_%s", tb.tableName, strings.Join(columns, "_"))
	index := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tb.tableName, strings.Join(columns, ", "))
	tb.indexes = append(tb.indexes, index)
}

// Timestamps adds timestamps
func (tb *TableBuilder) Timestamps() {
	driver := GetDriverName()
	tb.Column("created_at", "TIMESTAMP", "DEFAULT CURRENT_TIMESTAMP")
	if driver == "sqlite3" {
		tb.Column("updated_at", "TIMESTAMP", "DEFAULT CURRENT_TIMESTAMP")
	} else {
		tb.Column("updated_at", "TIMESTAMP", "DEFAULT CURRENT_TIMESTAMP", "ON UPDATE CURRENT_TIMESTAMP")
	}
}

// Build builds an SQL query
func (tb *TableBuilder) Build() (string, []string) {
	query := fmt.Sprintf("CREATE TABLE %s (\n", tb.tableName)
	query += strings.Join(tb.columns, ",\n")
	query += "\n)"
	return query, tb.indexes
}
