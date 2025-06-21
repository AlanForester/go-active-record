package activerecord

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Migration interface for migrations
type Migration interface {
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
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator() *Migrator {
	return &Migrator{db: GetConnection()}
}

// CreateMigrationsTable creates a table for tracking migrations
func (m *Migrator) CreateMigrationsTable() error {
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
func (m *Migrator) Migrate(migrations []Migration) error {
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
func (m *Migrator) Rollback(migrations []Migration) error {
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
	var targetMigration Migration
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
func (m *Migrator) Status(migrations []Migration) error {
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

func (m *Migrator) getAppliedMigrations() ([]MigrationRecord, error) {
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

func (m *Migrator) isMigrationApplied(applied []MigrationRecord, version int64) bool {
	for _, record := range applied {
		if record.Version == version {
			return true
		}
	}
	return false
}

func (m *Migrator) runMigration(migration Migration) error {
	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

func (m *Migrator) removeMigrationRecord(version int64) error {
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
