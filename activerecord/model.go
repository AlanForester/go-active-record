package activerecord

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
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
	switch v := id.(type) {
	case int64:
		m.ID = int(v)
	default:
		m.ID = id
	}
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

func setTimestampsDeep(val reflect.Value, now time.Time) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)
		if field.Kind() == reflect.Struct {
			setTimestampsDeep(field, now)
		}
		if fieldType.Name == "CreatedAt" && field.CanSet() {
			field.Set(reflect.ValueOf(now))
		}
		if fieldType.Name == "UpdatedAt" && field.CanSet() {
			field.Set(reflect.ValueOf(now))
		}
	}
}

func getFieldsAndValuesDeep(val reflect.Value, t reflect.Type, excludeID bool, fields *[]string, values *[]interface{}) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		if fieldType.Anonymous && (field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct)) {
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
	now := time.Now()
	setTimestampsDeep(reflect.ValueOf(model), now)

	// Try Modeler assertion
	var modeler Modeler
	if m, ok := model.(Modeler); ok {
		modeler = m
	} else {
		return fmt.Errorf("model does not implement Modeler, type: %T", model)
	}

	tableName := modeler.TableName()
	fields, values := getFieldsAndValues(model, true) // true = exclude ID

	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	db := GetConnection()
	driver := GetDriverName()

	var id interface{}
	if driver == "sqlite3" {
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "),
		)
		result, err := db.Exec(query, values...)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		id = lastID
	} else {
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
			tableName,
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "),
		)
		err := db.QueryRow(query, values...).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
	}

	modeler.SetID(id)
	// Reload the record from the DB to get all DB-managed fields
	if err := Find(modeler, modeler.GetID()); err != nil {
		return fmt.Errorf("failed to reload record after insert: %w", err)
	}
	return nil
}

// Find finds a record by ID
func Find(model interface{}, id interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return fmt.Errorf("model does not implement Modeler, type: %T", model)
	}
	tableName := modeler.TableName()
	fields, _ := getFieldsAndValues(model, false)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = ?",
		strings.Join(fields, ", "),
		tableName,
	)

	row := GetConnection().QueryRow(query, id)
	err := scanRow(row, model, fields)
	return err
}

// Update updates a record in the database
func Update(model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return fmt.Errorf("model does not implement Modeler, type: %T", model)
	}
	modeler.SetUpdatedAt(time.Now())

	tableName := modeler.TableName()
	fields, values := getFieldsAndValues(model, true) // exclude ID

	setClause := make([]string, len(fields))
	for i, field := range fields {
		setClause[i] = fmt.Sprintf("%s = ?", field)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		tableName,
		strings.Join(setClause, ", "),
	)

	values = append(values, modeler.GetID())

	_, err := GetConnection().Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete deletes a record from the database
func Delete(model interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return fmt.Errorf("model does not implement Modeler, type: %T", model)
	}
	tableName := modeler.TableName()
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	_, err := GetConnection().Exec(query, modeler.GetID())
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// FindAll finds all records
func FindAll(models interface{}) error {
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() != reflect.Ptr || modelsValue.Elem().Kind() != reflect.Slice {
		return errors.New("models must be a pointer to a slice")
	}

	sliceType := modelsValue.Elem().Type()
	elementType := sliceType.Elem()

	// Create an instance to get the table name
	instance := reflect.New(elementType).Interface()
	modeler, ok := instance.(Modeler)
	if !ok {
		return errors.New("model must implement Modeler interface")
	}
	tableName := modeler.TableName()
	fields, _ := getFieldsAndValues(instance, false)

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)

	rows, err := GetConnection().Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	slice := modelsValue.Elem()
	for rows.Next() {
		instance := reflect.New(elementType).Interface()
		if err := scanRow(rows, instance, fields); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(instance).Elem()))
	}

	return rows.Err()
}

// Where executes a query with conditions
func Where(models interface{}, query string, args ...interface{}) error {
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() != reflect.Ptr || modelsValue.Elem().Kind() != reflect.Slice {
		return errors.New("models must be a pointer to a slice")
	}

	sliceType := modelsValue.Elem().Type()
	elementType := sliceType.Elem()

	// Create an instance to get the table name
	instance := reflect.New(elementType).Interface()
	modeler, ok := instance.(Modeler)
	if !ok {
		return errors.New("model must implement Modeler interface")
	}
	tableName := modeler.TableName()
	fields, _ := getFieldsAndValues(instance, false)

	fullQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(fields, ", "),
		tableName,
		query,
	)

	rows, err := GetConnection().Query(fullQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	slice := modelsValue.Elem()
	for rows.Next() {
		instance := reflect.New(elementType).Interface()
		if err := scanRow(rows, instance, fields); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(instance).Elem()))
	}

	return rows.Err()
}

// Helper functions

func scanRow(row interface{}, model interface{}, fields []string) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	var scanArgs []interface{}
	var timeFieldIndexes []int
	var timeFieldPtrs []interface{}
	var timeFieldStructFields []reflect.Value

	for i, fieldName := range fields {
		field := val.FieldByNameFunc(func(name string) bool {
			fieldType, _ := typ.FieldByName(name)
			dbTag := fieldType.Tag.Get("db")
			if dbTag == "" {
				dbTag = strings.ToLower(name)
			}
			return dbTag == fieldName
		})
		if field.IsValid() && field.CanSet() {
			if field.Type() == reflect.TypeOf(time.Time{}) {
				var tmp string
				scanArgs = append(scanArgs, &tmp)
				timeFieldIndexes = append(timeFieldIndexes, i)
				timeFieldPtrs = append(timeFieldPtrs, &tmp)
				timeFieldStructFields = append(timeFieldStructFields, field)
			} else {
				scanArgs = append(scanArgs, field.Addr().Interface())
			}
		} else {
			var dummy interface{}
			scanArgs = append(scanArgs, &dummy)
		}
	}

	var err error
	switch r := row.(type) {
	case *sql.Row:
		err = r.Scan(scanArgs...)
	case *sql.Rows:
		err = r.Scan(scanArgs...)
	default:
		return errors.New("unsupported row type")
	}
	if err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	// Parse time fields after scan and set them in the struct
	for idx := range timeFieldIndexes {
		tmpPtr := timeFieldPtrs[idx].(*string)
		field := timeFieldStructFields[idx]
		if tmpPtr != nil && *tmpPtr != "" {
			t, _ := time.Parse(time.RFC3339, *tmpPtr)
			if t.IsZero() {
				t, _ = time.Parse("2006-01-02 15:04:05", *tmpPtr)
			}
			if field.CanSet() {
				field.Set(reflect.ValueOf(t))
			}
		}
	}
	return nil
}
