package activerecord

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// QueryInfo represents information about a database query
type QueryInfo struct {
	Query     string
	Args      []interface{}
	Duration  time.Duration
	Error     error
	Database  string
	Operation string
	Timestamp time.Time
}

// Logger interface for logging
type Logger interface {
	LogQuery(info *QueryInfo)
	Log(level LogLevel, message string, fields map[string]interface{})
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// DefaultLogger is the default logger implementation
type DefaultLogger struct {
	level    LogLevel
	queryLog *log.Logger
	appLog   *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:    InfoLevel,
		queryLog: log.New(os.Stdout, "[QUERY] ", log.LstdFlags),
		appLog:   log.New(os.Stdout, "[AR] ", log.LstdFlags),
	}
}

// LogQuery logs a database query
func (l *DefaultLogger) LogQuery(info *QueryInfo) {
	if l.level > InfoLevel {
		return
	}

	status := "SUCCESS"
	if info.Error != nil {
		status = "ERROR"
	}

	message := fmt.Sprintf("[%s] %s | %s | %v | %s",
		status,
		info.Operation,
		info.Duration,
		info.Query,
		info.Args,
	)

	if info.Database != "" {
		message = fmt.Sprintf("[%s] %s", info.Database, message)
	}

	l.queryLog.Println(message)

	if info.Error != nil {
		l.appLog.Printf("Query error: %v", info.Error)
	}
}

// Log logs a message with the specified level
func (l *DefaultLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	logMessage := fmt.Sprintf("[%s] %s", level.String(), message)
	if len(fields) > 0 {
		logMessage += fmt.Sprintf(" | %v", fields)
	}

	switch level {
	case DebugLevel, InfoLevel:
		l.appLog.Println(logMessage)
	case WarnLevel:
		l.appLog.Printf("WARNING: %s\n", logMessage)
	case ErrorLevel:
		l.appLog.Printf("ERROR: %s\n", logMessage)
	case FatalLevel:
		l.appLog.Printf("FATAL: %s\n", logMessage)
		os.Exit(1)
	}
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel returns the current logging level
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
}

// StructuredLogger is a structured logger implementation
type StructuredLogger struct {
	level    LogLevel
	queryLog *log.Logger
	appLog   *log.Logger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		level:    InfoLevel,
		queryLog: log.New(os.Stdout, "", log.LstdFlags),
		appLog:   log.New(os.Stdout, "", log.LstdFlags),
	}
}

// LogQuery logs a database query in structured format
func (sl *StructuredLogger) LogQuery(info *QueryInfo) {
	if sl.level > InfoLevel {
		return
	}

	status := "success"
	if info.Error != nil {
		status = "error"
	}

	fields := map[string]interface{}{
		"level":     "info",
		"component": "query",
		"status":    status,
		"operation": info.Operation,
		"duration":  info.Duration.String(),
		"query":     info.Query,
		"args":      info.Args,
		"timestamp": info.Timestamp.Format(time.RFC3339),
	}

	if info.Database != "" {
		fields["database"] = info.Database
	}

	if info.Error != nil {
		fields["error"] = info.Error.Error()
	}

	sl.logStructured("query", fields)
}

// Log logs a message in structured format
func (sl *StructuredLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	if level < sl.level {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["level"] = level.String()
	fields["message"] = message
	fields["timestamp"] = time.Now().Format(time.RFC3339)

	sl.logStructured("application", fields)
}

// logStructured logs in structured format
func (sl *StructuredLogger) logStructured(component string, fields map[string]interface{}) {
	fields["component"] = component

	// Simple JSON-like format
	logLine := fmt.Sprintf("{\"level\":\"%s\",\"component\":\"%s\",\"timestamp\":\"%s\"",
		fields["level"], component, fields["timestamp"])

	for key, value := range fields {
		if key != "level" && key != "component" && key != "timestamp" {
			logLine += fmt.Sprintf(",\"%s\":%v", key, value)
		}
	}

	logLine += "}"

	sl.appLog.Println(logLine)
}

// SetLevel sets the logging level
func (sl *StructuredLogger) SetLevel(level LogLevel) {
	sl.level = level
}

// GetLevel returns the current logging level
func (sl *StructuredLogger) GetLevel() LogLevel {
	return sl.level
}

// QueryLogger wraps database operations with logging
type QueryLogger struct {
	logger Logger
}

// NewQueryLogger creates a new query logger
func NewQueryLogger(logger Logger) *QueryLogger {
	return &QueryLogger{
		logger: logger,
	}
}

// LogExec logs an Exec operation
func (ql *QueryLogger) LogExec(query string, args []interface{}, start time.Time, result sql.Result, err error) {
	duration := time.Since(start)

	info := &QueryInfo{
		Query:     query,
		Args:      args,
		Duration:  duration,
		Error:     err,
		Operation: "EXEC",
		Timestamp: start,
	}

	ql.logger.LogQuery(info)
}

// LogQuery logs a Query operation
func (ql *QueryLogger) LogQuery(query string, args []interface{}, start time.Time, rows *sql.Rows, err error) {
	duration := time.Since(start)

	info := &QueryInfo{
		Query:     query,
		Args:      args,
		Duration:  duration,
		Error:     err,
		Operation: "QUERY",
		Timestamp: start,
	}

	ql.logger.LogQuery(info)
}

// LogQueryRow logs a QueryRow operation
func (ql *QueryLogger) LogQueryRow(query string, args []interface{}, start time.Time, row *sql.Row, err error) {
	duration := time.Since(start)

	info := &QueryInfo{
		Query:     query,
		Args:      args,
		Duration:  duration,
		Error:     err,
		Operation: "QUERY_ROW",
		Timestamp: start,
	}

	ql.logger.LogQuery(info)
}

// PerformanceMetrics tracks performance metrics
type PerformanceMetrics struct {
	TotalQueries       int64
	TotalDuration      time.Duration
	SlowQueries        int64
	SlowQueryThreshold time.Duration
	mu                 sync.RWMutex
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics(slowQueryThreshold time.Duration) *PerformanceMetrics {
	return &PerformanceMetrics{
		SlowQueryThreshold: slowQueryThreshold,
	}
}

// RecordQuery records a query for metrics
func (pm *PerformanceMetrics) RecordQuery(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.TotalQueries++
	pm.TotalDuration += duration

	if duration > pm.SlowQueryThreshold {
		pm.SlowQueries++
	}
}

// GetStats returns current statistics
func (pm *PerformanceMetrics) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var avgDuration time.Duration
	if pm.TotalQueries > 0 {
		avgDuration = pm.TotalDuration / time.Duration(pm.TotalQueries)
	}

	return map[string]interface{}{
		"total_queries":        pm.TotalQueries,
		"total_duration":       pm.TotalDuration.String(),
		"average_duration":     avgDuration.String(),
		"slow_queries":         pm.SlowQueries,
		"slow_query_threshold": pm.SlowQueryThreshold.String(),
	}
}

// Reset resets the metrics
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.TotalQueries = 0
	pm.TotalDuration = 0
	pm.SlowQueries = 0
}

// Global logger instance
var globalLogger Logger = NewDefaultLogger()
var globalQueryLogger *QueryLogger = NewQueryLogger(globalLogger)
var globalMetrics *PerformanceMetrics = NewPerformanceMetrics(100 * time.Millisecond)

// SetLogger sets the global logger
func SetLogger(logger Logger) {
	globalLogger = logger
	globalQueryLogger = NewQueryLogger(logger)
}

// GetLogger returns the global logger
func GetLogger() Logger {
	return globalLogger
}

// GetQueryLogger returns the global query logger
func GetQueryLogger() *QueryLogger {
	return globalQueryLogger
}

// GetMetrics returns the global performance metrics
func GetMetrics() *PerformanceMetrics {
	return globalMetrics
}

// LoggedExec executes a query with logging
func LoggedExec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := Exec(query, args...)

	GetQueryLogger().LogExec(query, args, start, result, err)
	GetMetrics().RecordQuery(time.Since(start))

	return result, err
}

// LoggedQuery executes a query with logging
func LoggedQuery(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := Query(query, args...)

	GetQueryLogger().LogQuery(query, args, start, rows, err)
	GetMetrics().RecordQuery(time.Since(start))

	return rows, err
}

// LoggedQueryRow executes a query row with logging
func LoggedQueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := QueryRow(query, args...)

	// Note: We can't get the error from sql.Row until Scan is called
	// So we'll log the query but not the error
	GetQueryLogger().LogQueryRow(query, args, start, row, nil)
	GetMetrics().RecordQuery(time.Since(start))

	return row
}

// LoggedExecWithContext executes a query with context and logging
func LoggedExecWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := GetConnection().ExecContext(ctx, query, args...)

	GetQueryLogger().LogExec(query, args, start, result, err)
	GetMetrics().RecordQuery(time.Since(start))

	return result, err
}

// LoggedQueryWithContext executes a query with context and logging
func LoggedQueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := GetConnection().QueryContext(ctx, query, args...)

	GetQueryLogger().LogQuery(query, args, start, rows, err)
	GetMetrics().RecordQuery(time.Since(start))

	return rows, err
}

// LoggedQueryRowWithContext executes a query row with context and logging
func LoggedQueryRowWithContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := GetConnection().QueryRowContext(ctx, query, args...)

	GetQueryLogger().LogQueryRow(query, args, start, row, nil)
	GetMetrics().RecordQuery(time.Since(start))

	return row
}

// Log logs a message using the global logger
func Log(level LogLevel, message string, fields map[string]interface{}) {
	globalLogger.Log(level, message, fields)
}

// LogDebug logs a debug message
func LogDebug(message string, fields map[string]interface{}) {
	Log(DebugLevel, message, fields)
}

// LogInfo logs an info message
func LogInfo(message string, fields map[string]interface{}) {
	Log(InfoLevel, message, fields)
}

// LogWarn logs a warning message
func LogWarn(message string, fields map[string]interface{}) {
	Log(WarnLevel, message, fields)
}

// LogError logs an error message
func LogError(message string, fields map[string]interface{}) {
	Log(ErrorLevel, message, fields)
}

// LogFatal logs a fatal message and exits
func LogFatal(message string, fields map[string]interface{}) {
	Log(FatalLevel, message, fields)
}

// GetPerformanceStats returns current performance statistics
func GetPerformanceStats() map[string]interface{} {
	return GetMetrics().GetStats()
}

// ResetPerformanceStats resets performance statistics
func ResetPerformanceStats() {
	GetMetrics().Reset()
}
