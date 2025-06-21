package activerecord

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Common errors.
var (
	ErrNotFound = errors.New("record not found")
)

// Model base structure for all Active Record models
type Model struct{}

// TableNamer interface for getting table name
type TableNamer interface {
	TableName() string
}

// Modeler interface for working with models
type Modeler interface {
	TableName() string
	GetID() interface{}
	SetID(interface{})
	GetCreatedAt() time.Time
	SetCreatedAt(time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(time.Time)
	Find(id interface{}) error
	Where(query string, args ...interface{}) (interface{}, error)
}

// BaseModel base model with common fields
type BaseModel struct {
	ID        interface{} `db:"id" json:"id"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

// GetID returns the ID of the model
func (m *BaseModel) GetID() interface{} {
	return m.ID
}

// SetID sets the ID of the model
func (m *BaseModel) SetID(id interface{}) {
	m.ID = id
}

// GetCreatedAt returns the creation time
func (m *BaseModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// SetCreatedAt sets the creation time
func (m *BaseModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}

// GetUpdatedAt returns the update time
func (m *BaseModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

// SetUpdatedAt sets the update time
func (m *BaseModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

// TableName returns the default table name
func (m *BaseModel) TableName() string {
	return "base_models"
}

// Find fills the receiver by id
func (m *BaseModel) Find(id interface{}) error {
	return Find(m, id)
}

// Where returns a slice of the receiver's type matching the query
func (m *BaseModel) Where(query string, args ...interface{}) (interface{}, error) {
	// This would need to be implemented based on the specific model type
	return []interface{}{nil}, fmt.Errorf("Where method not implemented for generic BaseModel")
}

func setTimestampsDeep(val reflect.Value, now time.Time) {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		switch {
		case fieldType.Anonymous && field.Kind() == reflect.Struct:
			setTimestampsDeep(field, now)
		case fieldType.Name == "CreatedAt" && field.CanSet():
			field.Set(reflect.ValueOf(now))
		case fieldType.Name == "UpdatedAt" && field.CanSet():
			field.Set(reflect.ValueOf(now))
		}
	}
}

func getFieldsAndValuesDeep(val reflect.Value, t reflect.Type, excludeID bool,
	fields *[]string, values *[]interface{}) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		if fieldType.Anonymous && (field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct)) {
			getFieldsAndValuesDeep(reflect.Indirect(field), reflect.Indirect(field).Type(), excludeID, fields, values)
			continue
		}
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(fieldType.Name)
		}
		if dbTag == "-" {
			continue
		}
		if excludeID && (dbTag == "id" || fieldType.Name == "ID") {
			continue
		}
		// Skip zero time values for created_at and updated_at
		if (dbTag == "created_at" || dbTag == "updated_at") && field.Type() == reflect.TypeOf(time.Time{}) {
			if field.Interface().(time.Time).IsZero() {
				continue
			}
		}
		*fields = append(*fields, dbTag)
		*values = append(*values, field.Interface())
	}
}

func getFieldsAndValues(model interface{}, excludeID bool) ([]string, []interface{}) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	var fields []string
	var values []interface{}
	getFieldsAndValuesDeep(val, val.Type(), excludeID, &fields, &values)
	return fields, values
}

// Create creates a new record in the database
func Create(model interface{}) error {
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

	// Execute query
	result, err := Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	// Set the generated ID
	if id, err := result.LastInsertId(); err == nil {
		modeler.SetID(id)
	}

	return nil
}

// Find finds a record by ID
func Find(model interface{}, id interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", modeler.TableName())
	row := QueryRow(query, id)
	if row == nil {
		return ErrNotFound
	}

	return scanRow(row, model)
}

// Update updates a record in the database
func Update(model interface{}) error {
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

	// Execute query
	_, err := Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete deletes a record from the database
func Delete(model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", modeler.TableName())
	_, err := Exec(query, modeler.GetID())
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// FindAll finds all records
func FindAll(models interface{}) error {
	val := reflect.ValueOf(models)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("models must be a pointer to a slice")
	}

	elemType := val.Elem().Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// Create a temporary instance to get table name
	temp := reflect.New(elemType).Interface()
	modeler, ok := temp.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	query := fmt.Sprintf("SELECT * FROM %s", modeler.TableName())
	rows, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	// Get column names.
	_, err = rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan rows.
	slice := val.Elem()
	for rows.Next() {
		// Create a new element.
		elem := reflect.New(elemType)
		if err := scanRow(rows, elem.Interface()); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		slice.Set(reflect.Append(slice, elem))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}

// Where fills the receiver slice with records matching the query.
func Where(models interface{}, query string, args ...interface{}) error {
	// Check if models is a pointer to a slice.
	val := reflect.ValueOf(models)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("models must be a pointer to a slice")
	}

	// Get the element type.
	elemType := val.Elem().Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// Create a temporary instance to get table name.
	temp := reflect.New(elemType).Interface()
	modeler, ok := temp.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	// Build the full query.
	fullQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s", modeler.TableName(), query)

	rows, err := Query(fullQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	// Get column names.
	_, err = rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan rows.
	slice := val.Elem()
	for rows.Next() {
		// Create a new element.
		elem := reflect.New(elemType)
		if err := scanRow(rows, elem.Interface()); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		slice.Set(reflect.Append(slice, elem))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}

// Helper functions

// scanRow scans a database row into a model.
func scanRow(row interface{}, model interface{}) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Рекурсивно собираем адреса всех полей (включая анонимные)
	var dest []interface{}
	collectFieldAddrs := func(v reflect.Value, t reflect.Type) {}
	collectFieldAddrs = func(v reflect.Value, t reflect.Type) {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)
			if fieldType.Anonymous && field.Kind() == reflect.Struct {
				collectFieldAddrs(field, field.Type())
				continue
			}
			dbTag := fieldType.Tag.Get("db")
			if dbTag == "-" {
				continue
			}
			dest = append(dest, field.Addr().Interface())
		}
	}
	collectFieldAddrs(val, val.Type())

	// Use the appropriate scan method based on the row type.
	switch r := row.(type) {
	case *sql.Row:
		return r.Scan(dest...)
	case *sql.Rows:
		return r.Scan(dest...)
	default:
		return fmt.Errorf("unsupported row type: %T", row)
	}
}
