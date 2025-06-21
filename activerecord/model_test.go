package activerecord

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestUser test user model
// Поля должны совпадать по порядку с таблицей test_users
// id, name, email, age, created_at, updated_at

type TestUser struct {
	ID        interface{} `db:"id"`
	Name      string      `db:"name" json:"name"`
	Email     string      `db:"email" json:"email"`
	Age       int         `db:"age" json:"age"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func (u *TestUser) TableName() string        { return "test_users" }
func (u *TestUser) GetID() interface{}       { return u.ID }
func (u *TestUser) SetID(id interface{})     { u.ID = id }
func (u *TestUser) GetCreatedAt() time.Time  { return u.CreatedAt }
func (u *TestUser) SetCreatedAt(t time.Time) { u.CreatedAt = t }
func (u *TestUser) GetUpdatedAt() time.Time  { return u.UpdatedAt }
func (u *TestUser) SetUpdatedAt(t time.Time) { u.UpdatedAt = t }

// TestValidationUser test model with validations
// id, name, email, age, created_at, updated_at

type TestValidationUser struct {
	ID               interface{} `db:"id"`
	Name             string      `db:"name" json:"name"`
	Email            string      `db:"email" json:"email"`
	Age              int         `db:"age" json:"age"`
	CreatedAt        time.Time   `db:"created_at"`
	UpdatedAt        time.Time   `db:"updated_at"`
	validationErrors ValidationErrors
	validationRules  []ValidationRule
}

func (u *TestValidationUser) TableName() string        { return "test_validation_users" }
func (u *TestValidationUser) GetID() interface{}       { return u.ID }
func (u *TestValidationUser) SetID(id interface{})     { u.ID = id }
func (u *TestValidationUser) GetCreatedAt() time.Time  { return u.CreatedAt }
func (u *TestValidationUser) SetCreatedAt(t time.Time) { u.CreatedAt = t }
func (u *TestValidationUser) GetUpdatedAt() time.Time  { return u.UpdatedAt }
func (u *TestValidationUser) SetUpdatedAt(t time.Time) { u.UpdatedAt = t }

func NewTestValidationUser() *TestValidationUser {
	return &TestValidationUser{
		validationErrors: ValidationErrors{},
		validationRules:  []ValidationRule{},
	}
}

// setupTestDB sets up the test database
func setupTestDB(t *testing.T) {
	// Use SQLite for tests
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set as global connection with driver name
	SetConnection(db, "sqlite3")

	// Create test table
	query := `
		CREATE TABLE test_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			age INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create table for validations
	query = `
		CREATE TABLE test_validation_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			age INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("Failed to create validation table: %v", err)
	}
}

// TestCreate tests record creation
func TestCreate(t *testing.T) {
	setupTestDB(t)
	user := &TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	if err := user.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.GetID() == nil {
		t.Error("ID should be set after create")
	}
	if user.GetCreatedAt().IsZero() {
		t.Error("CreatedAt should be set after create")
	}
	if user.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set after create")
	}
}

// TestFind tests finding a record by ID
func TestFind(t *testing.T) {
	setupTestDB(t)
	user := &TestUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}
	if err := user.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var foundUser TestUser
	if err := Find(&foundUser, user.GetID()); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}
	if foundUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
	}
}

// TestUpdate tests updating a record
func TestUpdate(t *testing.T) {
	setupTestDB(t)
	user := &TestUser{
		Name:  "Bob Smith",
		Email: "bob@example.com",
		Age:   35,
	}
	if err := user.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Name = "Bob Johnson"
	if err := user.Update(); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var foundUser TestUser
	if err := Find(&foundUser, user.GetID()); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if foundUser.Name != "Bob Johnson" {
		t.Errorf("Expected name Bob Johnson, got %s", foundUser.Name)
	}
}

// TestDelete tests deleting a record
func TestDelete(t *testing.T) {
	setupTestDB(t)
	user := &TestUser{
		Name:  "Alice Brown",
		Email: "alice@example.com",
		Age:   28,
	}
	if err := user.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := user.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var foundUser TestUser
	if err := Find(&foundUser, user.GetID()); err == nil {
		t.Error("Find should fail after delete")
	}
}

// TestFindAll tests finding all records
func TestFindAll(t *testing.T) {
	setupTestDB(t)
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25},
		{Name: "User2", Email: "user2@example.com", Age: 30},
		{Name: "User3", Email: "user3@example.com", Age: 35},
	}

	for i := range users {
		if err := users[i].Create(); err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
	}

	var foundUsers []*TestUser
	if err := FindAll(&foundUsers); err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}

// TestWhere tests querying with conditions
func TestWhere(t *testing.T) {
	setupTestDB(t)
	users := []*TestUser{
		{Name: "Young1", Email: "young1@example.com", Age: 20},
		{Name: "Young2", Email: "young2@example.com", Age: 22},
		{Name: "Old1", Email: "old1@example.com", Age: 40},
	}

	for i := range users {
		if err := users[i].Create(); err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
	}

	var youngUsers []*TestUser
	if err := Where(&youngUsers, "age < ?", 25); err != nil {
		t.Fatalf("Where failed: %v", err)
	}
	if len(youngUsers) != 2 {
		t.Errorf("Expected 2 young users, got %d", len(youngUsers))
	}
}

// TestValidations tests validations
func TestValidations(t *testing.T) {
	setupTestDB(t)

	user := NewTestValidationUser()
	user.Name = ""
	user.Email = "invalid-email"
	user.Age = 15

	// Set up validations
	user.PresenceOf("Name")
	user.AddValidation("Email", "email", "has invalid format")
	user.Numericality("Age", 18, 100)

	// Check validation
	if user.IsValid() {
		t.Error("Model should be invalid")
	}

	errors := user.Validate()
	if len(errors) == 0 {
		t.Error("Should have validation errors")
	}

	// Check valid model
	validUser := NewTestValidationUser()
	validUser.Name = "Valid User"
	validUser.Email = "valid@example.com"
	validUser.Age = 25
	validUser.PresenceOf("Name")
	validUser.AddValidation("Email", "email", "has invalid format")
	validUser.Numericality("Age", 18, 100)

	if !validUser.IsValid() {
		t.Error("Model should be valid")
	}
}

// TestIsNewRecord tests new record check
func TestIsNewRecord(t *testing.T) {
	user := &TestUser{}

	if !user.IsNewRecord() {
		t.Error("New model should be a new record")
	}

	user.SetID(1)
	if user.IsNewRecord() {
		t.Error("Model with ID should not be a new record")
	}
}

// TestIsPersisted tests persisted record check
func TestIsPersisted(t *testing.T) {
	user := &TestUser{}

	if user.IsPersisted() {
		t.Error("New model should not be persisted")
	}

	user.SetID(1)
	if !user.IsPersisted() {
		t.Error("Model with ID should be persisted")
	}
}

func (u *TestUser) Create() error {
	return Create(u)
}

func (u *TestUser) Update() error {
	return Update(u)
}

func (u *TestUser) Delete() error {
	return Delete(u)
}

func (u *TestUser) Reload() error {
	return Find(u, u.GetID())
}

// Modeler interface methods for TestUser
func (u *TestUser) Find(id interface{}) error {
	return Find(u, id)
}

func (u *TestUser) Where(query string, args ...interface{}) (interface{}, error) {
	var users []*TestUser
	err := Where(&users, query, args...)
	return users, err
}

// Modeler interface methods for TestValidationUser
func (u *TestValidationUser) Find(id interface{}) error {
	return Find(u, id)
}

func (u *TestValidationUser) Where(query string, args ...interface{}) (interface{}, error) {
	var users []*TestValidationUser
	err := Where(&users, query, args...)
	return users, err
}

func (u *TestUser) FindAll() ([]*TestUser, error) {
	var results []*TestUser
	err := FindAll(&results)
	return results, err
}

func (u *TestValidationUser) FindAll() ([]*TestValidationUser, error) {
	var results []*TestValidationUser
	err := FindAll(&results)
	return results, err
}

// Методы валидации для TestValidationUser
func (u *TestValidationUser) PresenceOf(field string) {
	u.AddValidation(field, "presence", field+" cannot be empty")
}
func (u *TestValidationUser) AddValidation(field, rule, message string, params ...interface{}) {
	u.validationRules = append(u.validationRules, ValidationRule{
		Field:   field,
		Rule:    rule,
		Message: message,
		Params:  params,
	})
}
func (u *TestValidationUser) Numericality(field string, min, max float64) {
	u.AddValidation(field, "numericality", field+" must be between range", min, max)
}
func (u *TestValidationUser) IsValid() bool {
	return len(u.Validate()) == 0
}
func (u *TestValidationUser) Validate() ValidationErrors {
	var errors ValidationErrors

	for _, rule := range u.validationRules {
		switch rule.Rule {
		case "presence":
			// Получаем значение поля через рефлексию
			val := reflect.ValueOf(u).Elem().FieldByName(rule.Field)
			if val.IsValid() && val.String() == "" {
				errors = append(errors, ValidationError{
					Field:   rule.Field,
					Message: rule.Message,
				})
			}
		case "email":
			val := reflect.ValueOf(u).Elem().FieldByName(rule.Field)
			if val.IsValid() {
				email := val.String()
				// Простая проверка email
				if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
					errors = append(errors, ValidationError{
						Field:   rule.Field,
						Message: rule.Message,
					})
				}
			}
		case "numericality":
			val := reflect.ValueOf(u).Elem().FieldByName(rule.Field)
			if val.IsValid() {
				age := val.Int()
				min := int64(rule.Params[0].(float64))
				max := int64(rule.Params[1].(float64))
				if age < min || age > max {
					errors = append(errors, ValidationError{
						Field:   rule.Field,
						Message: rule.Message,
					})
				}
			}
		}
	}

	return errors
}

// IsNewRecord для TestUser
func (u *TestUser) IsNewRecord() bool {
	return u.ID == nil || u.ID == 0
}

func (u *TestUser) IsPersisted() bool {
	return !u.IsNewRecord()
}
