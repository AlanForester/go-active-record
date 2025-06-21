package activerecord

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var globalDB *sql.DB
var globalDriverName string

// Connect establishes a connection to the database
func Connect(driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// SetConnection sets the global connection
func SetConnection(db *sql.DB, driverName ...string) {
	globalDB = db
	if len(driverName) > 0 {
		globalDriverName = driverName[0]
	}
}

// GetConnection returns the current connection
func GetConnection() *sql.DB {
	return globalDB
}

// GetDriverName returns the current driver name
func GetDriverName() string {
	return globalDriverName
}

// Close closes the connection to the database
func Close() error {
	if globalDB != nil {
		return globalDB.Close()
	}
	return nil
}

// Begin starts a transaction
func Begin() (*sql.Tx, error) {
	if globalDB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	return globalDB.Begin()
}

// Exec executes an SQL query without returning results
func Exec(query string, args ...interface{}) (sql.Result, error) {
	if globalDB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	return globalDB.Exec(query, args...)
}

// Query executes an SQL query and returns results
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	if globalDB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	return globalDB.Query(query, args...)
}

// QueryRow executes an SQL query and returns a single row
func QueryRow(query string, args ...interface{}) *sql.Row {
	if globalDB == nil {
		return nil
	}
	return globalDB.QueryRow(query, args...)
}
