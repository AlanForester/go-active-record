package activerecord

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// BatchInsertResult represents the result of a batch insert operation
type BatchInsertResult struct {
	LastInsertID int64
	RowsAffected int64
	Errors       []error
}

// BatchInsert performs batch insert of multiple records
func BatchInsert(models []interface{}) (*BatchInsertResult, error) {
	return BatchInsertWithContext(context.Background(), models)
}

// BatchInsertWithContext performs batch insert with context
func BatchInsertWithContext(ctx context.Context, models []interface{}) (*BatchInsertResult, error) {
	if len(models) == 0 {
		return &BatchInsertResult{}, nil
	}

	// Get the first model to determine structure
	firstModel := models[0]
	modeler, ok := firstModel.(Modeler)
	if !ok {
		return nil, ErrNotModeler
	}

	// Set timestamps on all models first
	now := time.Now()
	for _, model := range models {
		if m, ok := model.(Modeler); ok {
			m.SetCreatedAt(now)
			m.SetUpdatedAt(now)
		}
	}

	// Get fields and values for the first model (exclude ID)
	fields, values := getFieldsAndValues(firstModel, true)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to insert")
	}

	fmt.Printf("[DEBUG] BatchInsert: first model fields: %v\n", fields)
	fmt.Printf("[DEBUG] BatchInsert: first model values: %v\n", values)

	// Build the batch insert query
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	fmt.Printf("[DEBUG] BatchInsert: fields: %v\n", fields)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		modeler.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)
	fmt.Printf("[DEBUG] BatchInsert: query: %s\n", query)

	// Prepare the statement
	stmt, err := GetConnection().PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare batch insert statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch insert
	var lastInsertID int64
	var rowsAffected int64
	var errors []error

	for i, model := range models {
		// Get values for this model (exclude ID)
		_, values := getFieldsAndValues(model, true)
		fmt.Printf("[DEBUG] BatchInsert: inserting values: %v\n", values)

		// Execute insert
		result, err := stmt.ExecContext(ctx, values...)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to insert model at index %d: %w", i, err))
			continue
		}

		// Get last insert ID and rows affected
		if id, err := result.LastInsertId(); err == nil {
			lastInsertID = id
		}
		if affected, err := result.RowsAffected(); err == nil {
			rowsAffected += affected
		}

		// Set the generated ID on the model
		if m, ok := model.(Modeler); ok {
			m.SetID(lastInsertID)
		}
	}

	return &BatchInsertResult{
		LastInsertID: lastInsertID,
		RowsAffected: rowsAffected,
		Errors:       errors,
	}, nil
}

// BatchUpsert performs batch upsert (insert or update) operation
func BatchUpsert(models []interface{}, conflictFields []string, updateFields []string) (*BatchInsertResult, error) {
	return BatchUpsertWithContext(context.Background(), models, conflictFields, updateFields)
}

// BatchUpsertWithContext performs batch upsert with context
func BatchUpsertWithContext(ctx context.Context, models []interface{}, conflictFields []string, updateFields []string) (*BatchInsertResult, error) {
	if len(models) == 0 {
		return &BatchInsertResult{}, nil
	}

	// Get the first model to determine structure
	firstModel := models[0]
	modeler, ok := firstModel.(Modeler)
	if !ok {
		return nil, ErrNotModeler
	}

	// Get fields and values for the first model
	fields, _ := getFieldsAndValues(firstModel, false)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to insert")
	}

	// Build the batch upsert query
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		modeler.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	// Add ON CONFLICT clause for PostgreSQL
	if len(conflictFields) > 0 {
		query += fmt.Sprintf(" ON CONFLICT (%s)", strings.Join(conflictFields, ", "))

		if len(updateFields) > 0 {
			updateClauses := make([]string, len(updateFields))
			for i, field := range updateFields {
				updateClauses[i] = fmt.Sprintf("%s = EXCLUDED.%s", field, field)
			}
			query += fmt.Sprintf(" DO UPDATE SET %s", strings.Join(updateClauses, ", "))
		} else {
			query += " DO NOTHING"
		}
	}

	// Prepare the statement
	stmt, err := GetConnection().PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare batch upsert statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch upsert
	var lastInsertID int64
	var rowsAffected int64
	var errors []error

	for i, model := range models {
		// Get values for this model
		_, values := getFieldsAndValues(model, false)

		// Execute upsert
		result, err := stmt.ExecContext(ctx, values...)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to upsert model at index %d: %w", i, err))
			continue
		}

		// Get last insert ID and rows affected
		if id, err := result.LastInsertId(); err == nil {
			lastInsertID = id
		}
		if affected, err := result.RowsAffected(); err == nil {
			rowsAffected += affected
		}

		// Set the generated ID on the model
		if m, ok := model.(Modeler); ok {
			m.SetID(lastInsertID)
		}
	}

	return &BatchInsertResult{
		LastInsertID: lastInsertID,
		RowsAffected: rowsAffected,
		Errors:       errors,
	}, nil
}

// FindInBatches processes records in batches
func FindInBatches(modelType interface{}, batchSize int, fn func([]interface{}) error) error {
	return FindInBatchesWithContext(context.Background(), modelType, batchSize, fn)
}

// FindInBatchesWithContext processes records in batches with context
func FindInBatchesWithContext(ctx context.Context, modelType interface{}, batchSize int, fn func([]interface{}) error) error {
	// Get the type of the model
	modelTypeValue := reflect.TypeOf(modelType)
	if modelTypeValue.Kind() == reflect.Ptr {
		modelTypeValue = modelTypeValue.Elem()
	}

	// Create a temporary instance to get table name
	temp := reflect.New(modelTypeValue).Interface()
	modeler, ok := temp.(Modeler)
	if !ok {
		return fmt.Errorf("receiver does not implement Modeler")
	}

	offset := 0

	for {
		// Create query builder for this batch
		qb := NewQueryBuilder(modeler.TableName()).WithContext(ctx)
		qb.Limit(batchSize).Offset(offset)

		// Create slice for batch results
		batchType := reflect.SliceOf(reflect.PtrTo(modelTypeValue))
		batch := reflect.New(batchType).Interface()

		// Execute query
		err := qb.Find(batch)
		if err != nil {
			return fmt.Errorf("failed to find batch at offset %d: %w", offset, err)
		}

		// Convert to []interface{}
		batchSlice := reflect.ValueOf(batch).Elem()
		if batchSlice.Len() == 0 {
			break
		}
		batchInterfaces := make([]interface{}, batchSlice.Len())
		for i := 0; i < batchSlice.Len(); i++ {
			batchInterfaces[i] = batchSlice.Index(i).Interface()
		}

		// Call the callback function
		if err := fn(batchInterfaces); err != nil {
			return fmt.Errorf("batch callback failed at offset %d: %w", offset, err)
		}

		if len(batchInterfaces) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}

// FindOrCreate finds a record or creates it if it doesn't exist
func FindOrCreate(model interface{}, conditions map[string]interface{}) error {
	return FindOrCreateWithContext(context.Background(), model, conditions)
}

// FindOrCreateWithContext finds a record or creates it with context
func FindOrCreateWithContext(ctx context.Context, model interface{}, conditions map[string]interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	// Build where conditions
	var whereClauses []string
	var args []interface{}
	for field, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	if len(whereClauses) == 0 {
		return fmt.Errorf("no conditions provided for find or create")
	}

	// Try to find existing record
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1",
		modeler.TableName(),
		strings.Join(whereClauses, " AND "),
	)

	rows, err := GetConnection().QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query for existing record: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		// Record found, scan it
		if err := scanRow(rows, model); err != nil {
			return fmt.Errorf("failed to scan existing record: %w", err)
		}
		return nil
	}

	// Record not found, create it
	// Set the condition values on the model
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for field, value := range conditions {
		fieldVal := findFieldByTag(val, val.Type(), field)
		if fieldVal.IsValid() && fieldVal.CanSet() {
			if err := setFieldValue(fieldVal, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field, err)
			}
		}
	}

	// Create the record
	return Create(model)
}

// FindOrCreateByMap finds or creates records based on a map of attributes
func FindOrCreateByMap(modelType interface{}, attributes map[string]interface{}) (interface{}, error) {
	return FindOrCreateByMapWithContext(context.Background(), modelType, attributes)
}

// FindOrCreateByMapWithContext finds or creates records with context
func FindOrCreateByMapWithContext(ctx context.Context, modelType interface{}, attributes map[string]interface{}) (interface{}, error) {
	// Create a new instance of the model type
	model := reflect.New(reflect.TypeOf(modelType)).Interface()

	// Set attributes on the model
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for field, value := range attributes {
		fieldVal := findFieldByTag(val, val.Type(), field)
		if fieldVal.IsValid() && fieldVal.CanSet() {
			if err := setFieldValue(fieldVal, value); err != nil {
				return nil, fmt.Errorf("failed to set field %s: %w", field, err)
			}
		}
	}

	// Try to find or create
	err := FindOrCreateWithContext(ctx, model, attributes)
	return model, err
}

// UpdateWithSQLExpr updates a record using SQL expressions
func UpdateWithSQLExpr(model interface{}, expressions map[string]string, args ...interface{}) error {
	return UpdateWithSQLExprAndContext(context.Background(), model, expressions, args...)
}

// UpdateWithSQLExprAndContext updates a record using SQL expressions with context
func UpdateWithSQLExprAndContext(ctx context.Context, model interface{}, expressions map[string]string, args ...interface{}) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return ErrNotModeler
	}

	// Set updated timestamp
	modeler.SetUpdatedAt(time.Now())

	// Build SET clause with SQL expressions
	var setClauses []string
	for field, expr := range expressions {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", field, expr))
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no expressions provided for update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		modeler.TableName(),
		strings.Join(setClauses, ", "),
	)

	// Add ID to args
	allArgs := append(args, modeler.GetID())

	// Execute query
	_, err := GetConnection().ExecContext(ctx, query, allArgs...)
	if err != nil {
		return fmt.Errorf("failed to update record with SQL expressions: %w", err)
	}

	return nil
}

// DeleteWithConditions deletes records matching conditions
func DeleteWithConditions(modelType interface{}, conditions map[string]interface{}) (int64, error) {
	return DeleteWithConditionsAndContext(context.Background(), modelType, conditions)
}

// DeleteWithConditionsAndContext deletes records matching conditions with context
func DeleteWithConditionsAndContext(ctx context.Context, modelType interface{}, conditions map[string]interface{}) (int64, error) {
	// Create a temporary instance to get table name
	temp := reflect.New(reflect.TypeOf(modelType)).Interface()
	modeler, ok := temp.(Modeler)
	if !ok {
		return 0, ErrNotModeler
	}

	// Build where conditions
	var whereClauses []string
	var args []interface{}
	for field, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	if len(whereClauses) == 0 {
		return 0, fmt.Errorf("no conditions provided for delete")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		modeler.TableName(),
		strings.Join(whereClauses, " AND "),
	)

	result, err := GetConnection().ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BulkUpdate performs bulk update operation
func BulkUpdate(modelType interface{}, conditions map[string]interface{}, updates map[string]interface{}) (int64, error) {
	return BulkUpdateWithContext(context.Background(), modelType, conditions, updates)
}

// BulkUpdateWithContext performs bulk update with context
func BulkUpdateWithContext(ctx context.Context, modelType interface{}, conditions map[string]interface{}, updates map[string]interface{}) (int64, error) {
	// Create a temporary instance to get table name
	temp := reflect.New(reflect.TypeOf(modelType)).Interface()
	modeler, ok := temp.(Modeler)
	if !ok {
		return 0, ErrNotModeler
	}

	// Build SET clause
	var setClauses []string
	var setArgs []interface{}
	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		setArgs = append(setArgs, value)
	}

	if len(setClauses) == 0 {
		return 0, fmt.Errorf("no updates provided")
	}

	// Build WHERE clause
	var whereClauses []string
	var whereArgs []interface{}
	for field, value := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", field))
		whereArgs = append(whereArgs, value)
	}

	if len(whereClauses) == 0 {
		return 0, fmt.Errorf("no conditions provided")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		modeler.TableName(),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)

	// Combine args
	allArgs := append(setArgs, whereArgs...)

	result, err := GetConnection().ExecContext(ctx, query, allArgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
