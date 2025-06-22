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

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

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
	rows, err := Query(query, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}
	return scanRow(rows, model)
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

// Helper to recursively build a field map for all exported fields, including embedded structs
func buildFieldMap(val reflect.Value, typ reflect.Type, prefix string, fieldMap map[string]reflect.Value) {
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		field := val.Field(i)
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			buildFieldMap(field, field.Type(), prefix, fieldMap)
			continue
		}
		if !field.CanSet() {
			continue
		}
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(fieldType.Name)
		}
		if dbTag == "-" {
			continue
		}
		fieldMap[dbTag] = field
	}
}

// scanRow scans a database row into a model.
func scanRow(row interface{}, model interface{}) error {
	val := reflect.ValueOf(model)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("scanRow: model is not a struct")
	}
	typ := val.Type()

	columns, err := getColumns(row)
	if err != nil {
		return err
	}

	scanArgs := make([]interface{}, len(columns))
	fieldMap := make(map[string]reflect.Value)
	buildFieldMap(val, typ, "", fieldMap)

	for i, col := range columns {
		if field, ok := fieldMap[col]; ok {
			// Handle NULL values based on field type
			switch field.Kind() {
			case reflect.String:
				var nullStr sql.NullString
				scanArgs[i] = &nullStr
			case reflect.Int, reflect.Int64:
				var nullInt sql.NullInt64
				scanArgs[i] = &nullInt
			case reflect.Int32:
				var nullInt32 sql.NullInt32
				scanArgs[i] = &nullInt32
			case reflect.Int16:
				var nullInt16 sql.NullInt16
				scanArgs[i] = &nullInt16
			case reflect.Float64:
				var nullFloat sql.NullFloat64
				scanArgs[i] = &nullFloat
			case reflect.Bool:
				var nullBool sql.NullBool
				scanArgs[i] = &nullBool
			default:
				scanArgs[i] = field.Addr().Interface()
			}
		} else {
			var dummy interface{}
			scanArgs[i] = &dummy
		}
	}

	switch r := row.(type) {
	case *sql.Row:
		if err := r.Scan(scanArgs...); err != nil {
			return err
		}
	case *sql.Rows:
		if err := r.Scan(scanArgs...); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported row type: %T", row)
	}

	// Set the field values from the scanned data
	for i, col := range columns {
		if field, ok := fieldMap[col]; ok {
			switch field.Kind() {
			case reflect.String:
				if nullStr, ok := scanArgs[i].(*sql.NullString); ok && nullStr.Valid {
					field.SetString(nullStr.String)
				}
			case reflect.Int, reflect.Int64:
				if nullInt, ok := scanArgs[i].(*sql.NullInt64); ok && nullInt.Valid {
					field.SetInt(nullInt.Int64)
				}
			case reflect.Int32:
				if nullInt32, ok := scanArgs[i].(*sql.NullInt32); ok && nullInt32.Valid {
					field.SetInt(int64(nullInt32.Int32))
				}
			case reflect.Int16:
				if nullInt16, ok := scanArgs[i].(*sql.NullInt16); ok && nullInt16.Valid {
					field.SetInt(int64(nullInt16.Int16))
				}
			case reflect.Float64:
				if nullFloat, ok := scanArgs[i].(*sql.NullFloat64); ok && nullFloat.Valid {
					field.SetFloat(nullFloat.Float64)
				}
			case reflect.Bool:
				if nullBool, ok := scanArgs[i].(*sql.NullBool); ok && nullBool.Valid {
					field.SetBool(nullBool.Bool)
				}
			default:
				// For other types, the value was already set directly
			}
		}
	}

	return nil
}

func getColumns(row interface{}) ([]string, error) {
	switch r := row.(type) {
	case *sql.Row:
		// For *sql.Row, we need to use a different approach since we can't get columns directly
		// We'll use a temporary *sql.Rows to get the column information
		return nil, fmt.Errorf("cannot get columns from *sql.Row directly; use *sql.Rows for column info")
	case *sql.Rows:
		return r.Columns()
	default:
		return nil, fmt.Errorf("unsupported row type: %T", row)
	}
}
