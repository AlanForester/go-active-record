package activerecord

import (
	"database/sql"
	"testing"
	"time"
)

// TestUser test user model
type TestUser struct {
	ActiveRecordModel
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

// TableName returns the table name
func (u *TestUser) TableName() string {
	return "test_users"
}

// TestValidationUser test model with validations
type TestValidationUser struct {
	ValidationModel
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

func NewTestValidationUser() *TestValidationUser {
	return &TestValidationUser{
		ValidationModel: ValidationModel{
			validationErrors: ValidationErrors{},
			validationRules:  []ValidationRule{},
		},
	}
}

// TableName returns the table name
func (u *TestValidationUser) TableName() string {
	return "test_validation_users"
}

// setupTestDB sets up the test database
func setupTestDB(t *testing.T) *sql.DB {
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

	return db
}

// TestCreate tests record creation
func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	err := user.Create()
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	if user.GetID() == nil {
		t.Error("ID was not set after creation")
	}

	if user.GetCreatedAt().IsZero() {
		t.Error("CreatedAt was not set")
	}

	if user.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt was not set")
	}
}

// TestFind tests finding a record by ID
func TestFind(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create user
	user := &TestUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}
	user.Create()

	// Find user
	foundUser := &TestUser{}
	err := Find(foundUser, user.GetID())
	if err != nil {
		t.Errorf("Failed to find user: %v", err)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}
}

// TestUpdate tests updating a record
func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create user
	user := &TestUser{
		Name:  "Bob Smith",
		Email: "bob@example.com",
		Age:   35,
	}
	user.Create()

	// Update user
	user.Age = 36
	err := user.Update()
	if err != nil {
		t.Errorf("Failed to update user: %v", err)
	}

	// Check update
	foundUser := &TestUser{}
	Find(foundUser, user.GetID())

	if foundUser.Age != 36 {
		t.Errorf("Expected age 36, got %d", foundUser.Age)
	}
}

// TestDelete tests deleting a record
func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create user
	user := &TestUser{
		Name:  "Alice Johnson",
		Email: "alice@example.com",
		Age:   28,
	}
	user.Create()

	// Delete user
	err := user.Delete()
	if err != nil {
		t.Errorf("Failed to delete user: %v", err)
	}

	// Check that user is deleted
	foundUser := &TestUser{}
	err = Find(foundUser, user.GetID())
	if err == nil {
		t.Error("User should be deleted")
	}
}

// TestFindAll tests finding all records
func TestFindAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create several users
	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Age: 25},
		{Name: "User2", Email: "user2@example.com", Age: 30},
		{Name: "User3", Email: "user3@example.com", Age: 35},
	}

	for _, user := range users {
		user.Create()
	}

	// Find all users
	userModel := &TestUser{}
	foundUsers, err := userModel.FindAll()
	if err != nil {
		t.Errorf("Failed to find all users: %v", err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, found %d", len(foundUsers))
	}
}

// TestWhere tests querying with conditions
func TestWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create users of different ages
	users := []*TestUser{
		{Name: "Young1", Email: "young1@example.com", Age: 20},
		{Name: "Young2", Email: "young2@example.com", Age: 22},
		{Name: "Old1", Email: "old1@example.com", Age: 40},
		{Name: "Old2", Email: "old2@example.com", Age: 45},
	}

	for _, user := range users {
		user.Create()
	}

	// Find young users
	userModel := &TestUser{}
	youngUsers, err := userModel.Where("age < ?", 30)
	if err != nil {
		t.Errorf("Failed to find young users: %v", err)
	}

	if len(youngUsers) != 2 {
		t.Errorf("Expected 2 young users, found %d", len(youngUsers))
	}
}

// TestValidations tests validations
func TestValidations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := NewTestValidationUser()
	user.Name = ""
	user.Email = "invalid-email"
	user.Age = 15

	// Set up validations
	user.PresenceOf("Name")
	user.AddValidation("Email", "email", "has invalid format")
	user.Numericality("Age", 18, 100)

	// Check validation
	if user.IsValid(user) {
		t.Error("Model should be invalid")
	}

	errors := user.Validate(user)
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

	if !validUser.IsValid(validUser) {
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

func (u *TestUser) GetCreatedAt() time.Time  { return u.BaseModel.GetCreatedAt() }
func (u *TestUser) SetCreatedAt(t time.Time) { u.BaseModel.SetCreatedAt(t) }
func (u *TestUser) GetUpdatedAt() time.Time  { return u.BaseModel.GetUpdatedAt() }
func (u *TestUser) SetUpdatedAt(t time.Time) { u.BaseModel.SetUpdatedAt(t) }

func (u *TestValidationUser) Create() error {
	return Create(u)
}
func (u *TestValidationUser) Update() error {
	return Update(u)
}
func (u *TestValidationUser) Delete() error {
	return Delete(u)
}
func (u *TestValidationUser) Reload() error {
	return Find(u, u.GetID())
}
func (u *TestValidationUser) GetCreatedAt() time.Time  { return u.ValidationModel.GetCreatedAt() }
func (u *TestValidationUser) SetCreatedAt(t time.Time) { u.ValidationModel.SetCreatedAt(t) }
func (u *TestValidationUser) GetUpdatedAt() time.Time  { return u.ValidationModel.GetUpdatedAt() }
func (u *TestValidationUser) SetUpdatedAt(t time.Time) { u.ValidationModel.SetUpdatedAt(t) }

func (u *TestValidationUser) Outer() interface{} { return u }

func (u *TestUser) Find(id interface{}) error {
	return Find(u, id)
}

func (u *TestUser) Where(query string, args ...interface{}) ([]TestUser, error) {
	var results []TestUser
	err := Where(&results, query, args...)
	return results, err
}

func (u *TestValidationUser) Find(id interface{}) error {
	return Find(u, id)
}

func (u *TestValidationUser) Where(query string, args ...interface{}) ([]TestValidationUser, error) {
	var results []TestValidationUser
	err := Where(&results, query, args...)
	return results, err
}

func (u *TestUser) FindAll() ([]TestUser, error) {
	var results []TestUser
	err := FindAll(&results)
	return results, err
}

func (u *TestValidationUser) FindAll() ([]TestValidationUser, error) {
	var results []TestValidationUser
	err := FindAll(&results)
	return results, err
}
