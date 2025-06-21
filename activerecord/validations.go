package activerecord

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationRule правило validation
type ValidationRule struct {
	Field   string
	Rule    string
	Message string
	Params  []interface{}
}

// ValidationError ошибка validation
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors коллекция ошибок validation
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validatable интерфейс для валидируемых моделей
type Validatable interface {
	Validate() ValidationErrors
	IsValid() bool
	Errors() ValidationErrors
}

// ValidationModel базовая модель с validation
type ValidationModel struct {
	ActiveRecordModel
	validationErrors ValidationErrors
	validationRules  []ValidationRule
}

// Validate выполняет validation модели
func (m *ValidationModel) Validate(model interface{}) ValidationErrors {
	m.validationErrors = ValidationErrors{}
	for _, rule := range m.validationRules {
		if err := m.validateRule(model, rule); err != nil {
			m.validationErrors = append(m.validationErrors, *err)
		}
	}
	return m.validationErrors
}

// IsValid проверяет, валидна ли модель
func (m *ValidationModel) IsValid(model interface{}) bool {
	return len(m.Validate(model)) == 0
}

// Errors возвращает ошибки validation
func (m *ValidationModel) Errors() ValidationErrors {
	return m.validationErrors
}

// AddValidation добавляет правило validation
func (m *ValidationModel) AddValidation(field, rule string, message string, params ...interface{}) {
	m.validationRules = append(m.validationRules, ValidationRule{
		Field:   field,
		Rule:    rule,
		Message: message,
		Params:  params,
	})
}

// Валидаторы

// PresenceOf проверяет наличие значения
func (m *ValidationModel) PresenceOf(field string) {
	m.AddValidation(field, "presence", "cannot be empty")
}

// Length проверяет длину строки
func (m *ValidationModel) Length(field string, min, max int) {
	m.AddValidation(field, "length", fmt.Sprintf("must be between %d and %d characters", min, max), min, max)
}

// Email проверяет формат email
func (m *ValidationModel) Email(field string) {
	m.AddValidation(field, "email", "has invalid format")
}

// Uniqueness проверяет уникальность
func (m *ValidationModel) Uniqueness(field string) {
	m.AddValidation(field, "uniqueness", "must be unique")
}

// Numericality проверяет числовое значение
func (m *ValidationModel) Numericality(field string, min, max float64) {
	m.AddValidation(field, "numericality", fmt.Sprintf("must be between %f and %f", min, max), min, max)
}

// Format проверяет формат по регулярному выражению
func (m *ValidationModel) Format(field string, pattern string) {
	m.AddValidation(field, "format", "has invalid format", pattern)
}

// Вспомогательные методы

func (m *ValidationModel) validateRule(model interface{}, rule ValidationRule) *ValidationError {
	fieldValue := getFieldValue(model, rule.Field)

	switch rule.Rule {
	case "presence":
		if m.isEmpty(fieldValue) {
			return &ValidationError{Field: rule.Field, Message: rule.Message}
		}
	case "length":
		if str, ok := fieldValue.(string); ok {
			length := len(str)
			min := rule.Params[0].(int)
			max := rule.Params[1].(int)
			if length < min || length > max {
				return &ValidationError{Field: rule.Field, Message: rule.Message}
			}
		}
	case "email":
		if str, ok := fieldValue.(string); ok {
			if !m.isValidEmail(str) {
				return &ValidationError{Field: rule.Field, Message: rule.Message}
			}
		}
	case "uniqueness":
		if m.isDuplicate(rule.Field, fieldValue) {
			return &ValidationError{Field: rule.Field, Message: rule.Message}
		}
	case "numericality":
		if num, ok := m.toFloat(fieldValue); ok {
			min := rule.Params[0].(float64)
			max := rule.Params[1].(float64)
			if num < min || num > max {
				return &ValidationError{Field: rule.Field, Message: rule.Message}
			}
		}
	case "format":
		if str, ok := fieldValue.(string); ok {
			pattern := rule.Params[0].(string)
			if !m.matchesPattern(str, pattern) {
				return &ValidationError{Field: rule.Field, Message: rule.Message}
			}
		}
	}

	return nil
}

func getFieldValue(model interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// Try direct field
	field := val.FieldByName(fieldName)
	if field.IsValid() {
		return field.Interface()
	}
	// Recurse into embedded structs
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		ft := typ.Field(i)
		if ft.Anonymous && (f.Kind() == reflect.Struct || (f.Kind() == reflect.Ptr && f.Elem().Kind() == reflect.Struct)) {
			res := getFieldValue(reflect.Indirect(f).Interface(), fieldName)
			if res != nil {
				return res
			}
		}
	}
	return nil
}

// In ValidationModel
func (m *ValidationModel) getFieldValue(fieldName string) interface{} {
	// Use the outer struct if possible
	if mPtr, ok := any(m).(interface{ Outer() interface{} }); ok {
		return getFieldValue(mPtr.Outer(), fieldName)
	}
	return getFieldValue(m, fieldName)
}

func (m *ValidationModel) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []byte:
		return len(v) == 0
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

func (m *ValidationModel) isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func (m *ValidationModel) isDuplicate(field string, value interface{}) bool {
	// Here should be a uniqueness check in the database
	// Currently returns a stub
	return false
}

func (m *ValidationModel) toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

func (m *ValidationModel) matchesPattern(str, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}
