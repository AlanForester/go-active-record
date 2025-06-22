package activerecord

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// QueryMode represents the mode of query execution
type QueryMode int

const (
	NormalMode QueryMode = iota
	PreparedStatementMode
	DryRunMode
)

// QueryBuilder represents a query builder
type QueryBuilder struct {
	tableName    string
	selectFields []string
	whereClauses []string
	whereArgs    []interface{}
	joins        []string
	orderBy      []string
	groupBy      []string
	having       []string
	limit        int
	offset       int
	distinct     bool
	lock         string
	hints        []string
	mode         QueryMode
	ctx          context.Context
	preloads     []string
	includes     []string
	excludes     []string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(tableName string) *QueryBuilder {
	return &QueryBuilder{
		tableName:    tableName,
		selectFields: []string{"*"},
		whereClauses: make([]string, 0),
		whereArgs:    make([]interface{}, 0),
		joins:        make([]string, 0),
		orderBy:      make([]string, 0),
		groupBy:      make([]string, 0),
		having:       make([]string, 0),
		hints:        make([]string, 0),
		preloads:     make([]string, 0),
		includes:     make([]string, 0),
		excludes:     make([]string, 0),
		mode:         NormalMode,
		ctx:          context.Background(),
	}
}

// Select sets the fields to select
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = fields
	return qb
}

// Where adds a where clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

// WhereIn adds a where in clause
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb.Where("1 = 0") // Always false
	}

	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	return qb.Where(condition, values...)
}

// WhereNotIn adds a where not in clause
func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}

	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
	return qb.Where(condition, values...)
}

// WhereNull adds a where null clause
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NULL", field))
}

// WhereNotNull adds a where not null clause
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NOT NULL", field))
}

// Join adds a join clause
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	join := fmt.Sprintf("JOIN %s ON %s", table, condition)
	qb.joins = append(qb.joins, join)
	return qb
}

// LeftJoin adds a left join clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	join := fmt.Sprintf("LEFT JOIN %s ON %s", table, condition)
	qb.joins = append(qb.joins, join)
	return qb
}

// RightJoin adds a right join clause
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	join := fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition)
	qb.joins = append(qb.joins, join)
	return qb
}

// InnerJoin adds an inner join clause
func (qb *QueryBuilder) InnerJoin(table, condition string) *QueryBuilder {
	join := fmt.Sprintf("INNER JOIN %s ON %s", table, condition)
	qb.joins = append(qb.joins, join)
	return qb
}

// OrderBy adds an order by clause
func (qb *QueryBuilder) OrderBy(field, direction string) *QueryBuilder {
	order := fmt.Sprintf("%s %s", field, strings.ToUpper(direction))
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

// GroupBy adds a group by clause
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, fields...)
	return qb
}

// Having adds a having clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.having = append(qb.having, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

// Limit sets the limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Distinct sets distinct flag
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	return qb
}

// Lock adds a lock clause
func (qb *QueryBuilder) Lock(lock string) *QueryBuilder {
	qb.lock = lock
	return qb
}

// Hint adds a hint
func (qb *QueryBuilder) Hint(hint string) *QueryBuilder {
	qb.hints = append(qb.hints, hint)
	return qb
}

// SetMode sets the query mode
func (qb *QueryBuilder) SetMode(mode QueryMode) *QueryBuilder {
	qb.mode = mode
	return qb
}

// WithContext sets the context
func (qb *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	qb.ctx = ctx
	return qb
}

// Preload adds preload associations
func (qb *QueryBuilder) Preload(associations ...string) *QueryBuilder {
	qb.preloads = append(qb.preloads, associations...)
	return qb
}

// Include adds include associations
func (qb *QueryBuilder) Include(associations ...string) *QueryBuilder {
	qb.includes = append(qb.includes, associations...)
	return qb
}

// Exclude adds exclude fields
func (qb *QueryBuilder) Exclude(fields ...string) *QueryBuilder {
	qb.excludes = append(qb.excludes, fields...)
	return qb
}

// Build builds the SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// Add hints if any
	if len(qb.hints) > 0 {
		query.WriteString(strings.Join(qb.hints, " ") + " ")
	}

	// SELECT
	query.WriteString("SELECT ")
	if qb.distinct {
		query.WriteString("DISTINCT ")
	}
	query.WriteString(strings.Join(qb.selectFields, ", "))

	// FROM
	query.WriteString(" FROM ")
	query.WriteString(qb.tableName)

	// JOINS
	if len(qb.joins) > 0 {
		query.WriteString(" ")
		query.WriteString(strings.Join(qb.joins, " "))
	}

	// WHERE
	if len(qb.whereClauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereClauses, " AND "))
	}

	// GROUP BY
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		query.WriteString(strings.Join(qb.having, " AND "))
	}

	// ORDER BY
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// LIMIT
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	// LOCK
	if qb.lock != "" {
		query.WriteString(" ")
		query.WriteString(qb.lock)
	}

	return query.String(), qb.whereArgs
}

// Execute executes the query and returns rows
func (qb *QueryBuilder) Execute() (*sql.Rows, error) {
	query, args := qb.Build()

	if qb.mode == DryRunMode {
		fmt.Printf("DRY RUN - Query: %s, Args: %v\n", query, args)
		return nil, nil
	}

	if qb.mode == PreparedStatementMode {
		stmt, err := GetConnection().PrepareContext(qb.ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()
		return stmt.QueryContext(qb.ctx, args...)
	}

	return GetConnection().QueryContext(qb.ctx, query, args...)
}

// Find executes the query and scans results into the provided slice
func (qb *QueryBuilder) Find(models interface{}) error {
	rows, err := qb.Execute()
	if err != nil {
		return err
	}
	if rows == nil {
		return nil // Dry run mode
	}
	defer rows.Close()

	return scanRows(rows, models)
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First(model interface{}) error {
	qb.Limit(1)
	rows, err := qb.Execute()
	if err != nil {
		return err
	}
	if rows == nil {
		return nil // Dry run mode
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}

	return scanRow(rows, model)
}

// Count executes a count query
func (qb *QueryBuilder) Count() (int64, error) {
	originalSelect := qb.selectFields
	qb.selectFields = []string{"COUNT(*)"}

	query, args := qb.Build()

	if qb.mode == DryRunMode {
		fmt.Printf("DRY RUN - Count Query: %s, Args: %v\n", query, args)
		return 0, nil
	}

	var count int64
	err := GetConnection().QueryRowContext(qb.ctx, query, args...).Scan(&count)

	// Restore original select fields
	qb.selectFields = originalSelect

	return count, err
}

// Exists checks if any records exist
func (qb *QueryBuilder) Exists() (bool, error) {
	qb.Limit(1)
	rows, err := qb.Execute()
	if err != nil {
		return false, err
	}
	if rows == nil {
		return false, nil // Dry run mode
	}
	defer rows.Close()

	return rows.Next(), nil
}

// Pluck executes the query and returns a slice of values from a single column
func (qb *QueryBuilder) Pluck(column string, values interface{}) error {
	originalSelect := qb.selectFields
	qb.selectFields = []string{column}

	rows, err := qb.Execute()
	if err != nil {
		return err
	}
	if rows == nil {
		return nil // Dry run mode
	}
	defer rows.Close()

	// Restore original select fields
	qb.selectFields = originalSelect

	return scanColumn(rows, values)
}

// Batch processing methods

// FindInBatches processes records in batches
func (qb *QueryBuilder) FindInBatches(batchSize int, fn func([]interface{}) error) error {
	offset := 0

	for {
		batchQB := qb.clone()
		batchQB.Limit(batchSize).Offset(offset)

		var batch []interface{}
		err := batchQB.Find(&batch)
		if err != nil {
			return err
		}

		if len(batch) == 0 {
			break
		}

		if err := fn(batch); err != nil {
			return err
		}

		if len(batch) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}

// clone creates a copy of the query builder
func (qb *QueryBuilder) clone() *QueryBuilder {
	return &QueryBuilder{
		tableName:    qb.tableName,
		selectFields: append([]string{}, qb.selectFields...),
		whereClauses: append([]string{}, qb.whereClauses...),
		whereArgs:    append([]interface{}{}, qb.whereArgs...),
		joins:        append([]string{}, qb.joins...),
		orderBy:      append([]string{}, qb.orderBy...),
		groupBy:      append([]string{}, qb.groupBy...),
		having:       append([]string{}, qb.having...),
		limit:        qb.limit,
		offset:       qb.offset,
		distinct:     qb.distinct,
		lock:         qb.lock,
		hints:        append([]string{}, qb.hints...),
		mode:         qb.mode,
		ctx:          qb.ctx,
		preloads:     append([]string{}, qb.preloads...),
		includes:     append([]string{}, qb.includes...),
		excludes:     append([]string{}, qb.excludes...),
	}
}

// Helper functions for scanning

func scanRows(rows *sql.Rows, models interface{}) error {
	val := reflect.ValueOf(models)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("models must be a pointer to a slice")
	}

	slice := val.Elem()
	elementType := slice.Type().Elem()
	isPtr := elementType.Kind() == reflect.Ptr
	structType := elementType
	if isPtr {
		structType = elementType.Elem()
	}

	for rows.Next() {
		elementPtr := reflect.New(structType) // always a pointer to struct
		fmt.Printf("[DEBUG] scanRows: passing type %T to scanRow\n", elementPtr.Interface())
		if err := scanRow(rows, elementPtr.Interface()); err != nil {
			return err
		}
		if isPtr {
			slice.Set(reflect.Append(slice, elementPtr))
		} else {
			slice.Set(reflect.Append(slice, elementPtr.Elem()))
		}
	}

	return rows.Err()
}

func scanColumn(rows *sql.Rows, values interface{}) error {
	val := reflect.ValueOf(values)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("values must be a pointer to a slice")
	}

	slice := val.Elem()

	for rows.Next() {
		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, reflect.ValueOf(value)))
	}

	return rows.Err()
}

func mapRowToModel(columns []string, values []interface{}, model interface{}) error {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i, column := range columns {
		field := findFieldByTag(val, typ, column)
		if field.IsValid() && field.CanSet() {
			value := values[i]
			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", column, err)
			}
		}
	}

	return nil
}

func findFieldByTag(val reflect.Value, typ reflect.Type, tag string) reflect.Value {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		dbTag := fieldType.Tag.Get("db")
		if dbTag == tag {
			return field
		}

		// Check embedded structs
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			if nestedField := findFieldByTag(field, field.Type(), tag); nestedField.IsValid() {
				return nestedField
			}
		}
	}

	return reflect.Value{}
}

func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	val := reflect.ValueOf(value)

	// Handle time.Time conversion
	if field.Type() == reflect.TypeOf(time.Time{}) {
		if timeVal, ok := value.(time.Time); ok {
			field.Set(reflect.ValueOf(timeVal))
			return nil
		}
		// Try to parse string as time
		if strVal, ok := value.(string); ok {
			if t, err := time.Parse(time.RFC3339, strVal); err == nil {
				field.Set(reflect.ValueOf(t))
				return nil
			}
		}
	}

	// Handle basic type conversions
	if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert %v to %v", val.Type(), field.Type())
}
