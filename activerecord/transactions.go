package activerecord

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Transaction represents a database transaction
type Transaction struct {
	tx          *sql.Tx
	ctx         context.Context
	savepoints  []string
	savepointID int
	mu          sync.Mutex
	committed   bool
	rolledBack  bool
	parentTx    *Transaction
	callbacks   []func() error
}

// TransactionManager manages transactions
type TransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin() (*Transaction, error) {
	return tm.BeginWithContext(context.Background())
}

// BeginWithContext starts a new transaction with context
func (tm *TransactionManager) BeginWithContext(ctx context.Context) (*Transaction, error) {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:          tx,
		ctx:         ctx,
		savepoints:  make([]string, 0),
		savepointID: 0,
	}, nil
}

// BeginNested starts a nested transaction (savepoint)
func (t *Transaction) BeginNested() (*Transaction, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return nil, fmt.Errorf("cannot begin nested transaction on committed/rolled back transaction")
	}

	t.savepointID++
	savepointName := fmt.Sprintf("sp_%d", t.savepointID)

	_, err := t.tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepointName))
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	t.savepoints = append(t.savepoints, savepointName)

	return &Transaction{
		tx:          t.tx,
		ctx:         t.ctx,
		savepoints:  make([]string, 0),
		savepointID: 0,
		parentTx:    t,
	}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed {
		return fmt.Errorf("transaction already committed")
	}

	if t.rolledBack {
		return fmt.Errorf("cannot commit rolled back transaction")
	}

	// If this is a nested transaction, just release the savepoint
	if t.parentTx != nil {
		if len(t.savepoints) > 0 {
			savepointName := t.savepoints[len(t.savepoints)-1]
			_, err := t.tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName))
			if err != nil {
				return fmt.Errorf("failed to release savepoint %s: %w", savepointName, err)
			}
		}
		t.committed = true
		return nil
	}

	// Run callbacks before commit
	for _, callback := range t.callbacks {
		if err := callback(); err != nil {
			return fmt.Errorf("commit callback failed: %w", err)
		}
	}

	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.committed = true
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed {
		return fmt.Errorf("cannot rollback committed transaction")
	}

	if t.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	// If this is a nested transaction, rollback to the savepoint
	if t.parentTx != nil {
		if len(t.savepoints) > 0 {
			savepointName := t.savepoints[len(t.savepoints)-1]
			_, err := t.tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName))
			if err != nil {
				return fmt.Errorf("failed to rollback to savepoint %s: %w", savepointName, err)
			}
		}
		t.rolledBack = true
		return nil
	}

	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	t.rolledBack = true
	return nil
}

// RollbackToSavepoint rolls back to a specific savepoint
func (t *Transaction) RollbackToSavepoint(savepointName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("cannot rollback to savepoint on committed/rolled back transaction")
	}

	// Check if savepoint exists
	found := false
	for _, sp := range t.savepoints {
		if sp == savepointName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("savepoint %s not found", savepointName)
	}

	_, err := t.tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", savepointName, err)
	}

	// Remove savepoints after the rollback point
	for i, sp := range t.savepoints {
		if sp == savepointName {
			t.savepoints = t.savepoints[:i+1]
			break
		}
	}

	return nil
}

// CreateSavepoint creates a named savepoint
func (t *Transaction) CreateSavepoint(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("cannot create savepoint on committed/rolled back transaction")
	}

	_, err := t.tx.Exec(fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}

	t.savepoints = append(t.savepoints, name)
	return nil
}

// ReleaseSavepoint releases a named savepoint
func (t *Transaction) ReleaseSavepoint(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.committed || t.rolledBack {
		return fmt.Errorf("cannot release savepoint on committed/rolled back transaction")
	}

	_, err := t.tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", name, err)
	}

	// Remove savepoint from list
	for i, sp := range t.savepoints {
		if sp == name {
			t.savepoints = append(t.savepoints[:i], t.savepoints[i+1:]...)
			break
		}
	}

	return nil
}

// AddCallback adds a callback to be executed before commit
func (t *Transaction) AddCallback(callback func() error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.callbacks = append(t.callbacks, callback)
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

// Query executes a query and returns rows
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

// QueryRow executes a query and returns a single row
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

// IsCommitted returns true if the transaction is committed
func (t *Transaction) IsCommitted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.committed
}

// IsRolledBack returns true if the transaction is rolled back
func (t *Transaction) IsRolledBack() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rolledBack
}

// GetSavepoints returns the list of savepoints
func (t *Transaction) GetSavepoints() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string{}, t.savepoints...)
}

// Transactional executes a function within a transaction
func Transactional(fn func(*Transaction) error) error {
	tx, err := GetConnection().Begin()
	if err != nil {
		return err
	}

	transaction := &Transaction{
		tx:          tx,
		ctx:         context.Background(),
		savepoints:  make([]string, 0),
		savepointID: 0,
	}

	defer func() {
		if !transaction.IsCommitted() && !transaction.IsRolledBack() {
			transaction.Rollback()
		}
	}()

	if err := fn(transaction); err != nil {
		return err
	}

	return transaction.Commit()
}

// TransactionalWithContext executes a function within a transaction with context
func TransactionalWithContext(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	transaction := &Transaction{
		tx:          tx,
		ctx:         ctx,
		savepoints:  make([]string, 0),
		savepointID: 0,
	}

	defer func() {
		if !transaction.IsCommitted() && !transaction.IsRolledBack() {
			transaction.Rollback()
		}
	}()

	if err := fn(transaction); err != nil {
		return err
	}

	return transaction.Commit()
}

// Global transaction manager instance
var globalTransactionManager *TransactionManager

// SetTransactionManager sets the global transaction manager
func SetTransactionManager(tm *TransactionManager) {
	globalTransactionManager = tm
}

// GetTransactionManager returns the global transaction manager
func GetTransactionManager() *TransactionManager {
	if globalTransactionManager == nil {
		globalTransactionManager = NewTransactionManager(GetConnection())
	}
	return globalTransactionManager
}
