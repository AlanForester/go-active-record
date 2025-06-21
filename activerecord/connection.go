package activerecord

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db         *sql.DB
	driverName string
)

// Connect establishes a connection to the database.
func Connect(driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	SetConnection(db, driver)
	return db, nil
}

// SetConnection sets the global connection.
func SetConnection(database *sql.DB, driver string) {
	db = database
	driverName = driver
}

// GetConnection returns the current connection.
func GetConnection() *sql.DB {
	return db
}

// GetDriverName returns the current driver name.
func GetDriverName() string {
	return driverName
}

// Close closes the connection to the database.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Begin starts a transaction.
func Begin() (*sql.Tx, error) {
	if db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return db.Begin()
}

// Exec executes an SQL query without returning results.
func Exec(query string, args ...interface{}) (sql.Result, error) {
	if db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return db.Exec(query, args...)
}

// Query executes an SQL query and returns results.
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	if db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return db.Query(query, args...)
}

// QueryRow executes an SQL query and returns a single row.
func QueryRow(query string, args ...interface{}) *sql.Row {
	if db == nil {
		log.Printf("Warning: no database connection")
		return nil
	}
	return db.QueryRow(query, args...)
}
