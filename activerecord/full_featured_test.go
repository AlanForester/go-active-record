package activerecord

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// Test models for demonstrating features

type FullFeaturedUser struct {
	HookableModel
	Name     string               `db:"name" json:"name"`
	Email    string               `db:"email" json:"email"`
	Age      int                  `db:"age" json:"age"`
	Posts    []*FullFeaturedPost  `db:"-" json:"posts"`
	Profile  *FullFeaturedProfile `db:"-" json:"profile"`
	Database string               `db:"database" json:"database"`
}

func (u *FullFeaturedUser) TableName() string {
	return "full_featured_users"
}

func (u *FullFeaturedUser) SetupHooks() {
	// Add hooks for demonstration
	u.AddHook(BeforeCreate, func(model interface{}) error {
		if user, ok := model.(*FullFeaturedUser); ok {
			log.Printf("Before creating user: %s", user.Name)
		}
		return nil
	})

	u.AddHook(AfterCreate, func(model interface{}) error {
		if user, ok := model.(*FullFeaturedUser); ok {
			log.Printf("After creating user: %s with ID: %v", user.Name, user.GetID())
		}
		return nil
	})

	u.AddHook(BeforeUpdate, func(model interface{}) error {
		if user, ok := model.(*FullFeaturedUser); ok {
			log.Printf("Before updating user: %s", user.Name)
		}
		return nil
	})

	u.AddHook(AfterUpdate, func(model interface{}) error {
		if user, ok := model.(*FullFeaturedUser); ok {
			log.Printf("After updating user: %s", user.Name)
		}
		return nil
	})
}

func (u *FullFeaturedUser) Create() error {
	if err := u.RunHooks(BeforeCreate); err != nil {
		return err
	}
	if err := Create(u); err != nil {
		return err
	}
	return u.RunHooks(AfterCreate)
}

func (u *FullFeaturedUser) Update() error {
	if err := u.RunHooks(BeforeUpdate); err != nil {
		return err
	}
	if err := Update(u); err != nil {
		return err
	}
	return u.RunHooks(AfterUpdate)
}

func (u *FullFeaturedUser) Delete() error {
	if err := u.RunHooks(BeforeDelete); err != nil {
		return err
	}
	if err := Delete(u); err != nil {
		return err
	}
	return u.RunHooks(AfterDelete)
}

func (u *FullFeaturedUser) Save() error {
	if err := u.RunHooks(BeforeSave); err != nil {
		return err
	}
	var err error
	if u.IsNewRecord() {
		err = u.Create()
	} else {
		err = u.Update()
	}
	if err != nil {
		return err
	}
	return u.RunHooks(AfterSave)
}

type FullFeaturedPost struct {
	HookableModel
	Title    string            `db:"title" json:"title"`
	Content  string            `db:"content" json:"content"`
	UserID   int               `db:"user_id" json:"user_id"`
	User     *FullFeaturedUser `db:"-" json:"user"`
	Database string            `db:"database" json:"database"`
}

func (p *FullFeaturedPost) TableName() string {
	return "full_featured_posts"
}

func (p *FullFeaturedPost) Create() error {
	if err := p.RunHooks(BeforeCreate); err != nil {
		return err
	}
	if err := Create(p); err != nil {
		return err
	}
	return p.RunHooks(AfterCreate)
}

func (p *FullFeaturedPost) Update() error {
	if err := p.RunHooks(BeforeUpdate); err != nil {
		return err
	}
	if err := Update(p); err != nil {
		return err
	}
	return p.RunHooks(AfterUpdate)
}

func (p *FullFeaturedPost) Delete() error {
	if err := p.RunHooks(BeforeDelete); err != nil {
		return err
	}
	if err := Delete(p); err != nil {
		return err
	}
	return p.RunHooks(AfterDelete)
}

func (p *FullFeaturedPost) Save() error {
	if err := p.RunHooks(BeforeSave); err != nil {
		return err
	}
	var err error
	if p.IsNewRecord() {
		err = p.Create()
	} else {
		err = p.Update()
	}
	if err != nil {
		return err
	}
	return p.RunHooks(AfterSave)
}

type FullFeaturedProfile struct {
	HookableModel
	Bio      string            `db:"bio" json:"bio"`
	UserID   int               `db:"user_id" json:"user_id"`
	User     *FullFeaturedUser `db:"-" json:"user"`
	Database string            `db:"database" json:"database"`
}

func (p *FullFeaturedProfile) TableName() string {
	return "full_featured_profiles"
}

func (p *FullFeaturedProfile) Create() error {
	if err := p.RunHooks(BeforeCreate); err != nil {
		return err
	}
	if err := Create(p); err != nil {
		return err
	}
	return p.RunHooks(AfterCreate)
}

func (p *FullFeaturedProfile) Update() error {
	if err := p.RunHooks(BeforeUpdate); err != nil {
		return err
	}
	if err := Update(p); err != nil {
		return err
	}
	return p.RunHooks(AfterUpdate)
}

func (p *FullFeaturedProfile) Delete() error {
	if err := p.RunHooks(BeforeDelete); err != nil {
		return err
	}
	if err := Delete(p); err != nil {
		return err
	}
	return p.RunHooks(AfterDelete)
}

func (p *FullFeaturedProfile) Save() error {
	if err := p.RunHooks(BeforeSave); err != nil {
		return err
	}
	var err error
	if p.IsNewRecord() {
		err = p.Create()
	} else {
		err = p.Update()
	}
	if err != nil {
		return err
	}
	return p.RunHooks(AfterSave)
}

// TestFullFeaturedORM demonstrates all the full-featured ORM capabilities
func TestFullFeaturedORM(t *testing.T) {
	// Setup database connection
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup logging
	logger := NewStructuredLogger()
	logger.SetLevel(DebugLevel)
	SetLogger(logger)

	// Create tables
	createTables(t)

	// Test 1: Basic CRUD with hooks
	t.Run("Basic CRUD with Hooks", func(t *testing.T) {
		testBasicCRUDWithHooks(t)
	})

	// Test 2: Associations
	t.Run("Associations", func(t *testing.T) {
		testAssociations(t)
	})

	// Test 3: Query Builder
	t.Run("Query Builder", func(t *testing.T) {
		testQueryBuilder(t)
	})

	// Test 4: Transactions
	t.Run("Transactions", func(t *testing.T) {
		testTransactions(t)
	})

	// Test 5: Batch Operations
	t.Run("Batch Operations", func(t *testing.T) {
		testBatchOperations(t)
	})

	// Test 6: Database Resolver
	t.Run("Database Resolver", func(t *testing.T) {
		testDatabaseResolver(t)
	})

	// Test 7: Logging and Performance
	t.Run("Logging and Performance", func(t *testing.T) {
		testLoggingAndPerformance(t)
	})
}

func createTables(t *testing.T) {
	// Create full_featured_users table
	_, err := Exec(`
		CREATE TABLE IF NOT EXISTS full_featured_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER,
			database TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create full_featured_users table: %v", err)
	}

	// Create full_featured_posts table
	_, err = Exec(`
		CREATE TABLE IF NOT EXISTS full_featured_posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT,
			user_id INTEGER,
			database TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES full_featured_users (id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create full_featured_posts table: %v", err)
	}

	// Create full_featured_profiles table
	_, err = Exec(`
		CREATE TABLE IF NOT EXISTS full_featured_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bio TEXT,
			user_id INTEGER UNIQUE,
			database TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES full_featured_users (id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create full_featured_profiles table: %v", err)
	}
}

func testBasicCRUDWithHooks(t *testing.T) {
	// Create user with hooks
	user := &FullFeaturedUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	user.SetupHooks()

	// Debug: Check table name
	tableName := user.TableName()
	t.Logf("Table name: %s", tableName)

	// Create
	err := user.Create()
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.GetID() == nil {
		t.Error("User ID should be set after creation")
	}

	// Find
	foundUser := &FullFeaturedUser{}
	err = Find(foundUser, user.GetID())
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	// Debug: Check if ID is set
	t.Logf("Found user ID: %v, type: %T", foundUser.GetID(), foundUser.GetID())

	// Update
	foundUser.Age = 31
	err = foundUser.Update()
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	updatedUser := &FullFeaturedUser{}
	err = Find(updatedUser, user.GetID())
	if err != nil {
		t.Fatalf("Failed to find updated user: %v", err)
	}

	if updatedUser.Age != 31 {
		t.Errorf("Expected age 31, got %d", updatedUser.Age)
	}

	// Delete
	err = updatedUser.Delete()
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion
	err = Find(&FullFeaturedUser{}, user.GetID())
	if err == nil {
		t.Error("User should not be found after deletion")
	}
}

func testAssociations(t *testing.T) {
	// Create user
	user := &FullFeaturedUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}
	err := user.Create()
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create posts for user
	userID := user.GetID().(int64)
	post1 := &FullFeaturedPost{
		Title:   "First Post",
		Content: "This is the first post",
		UserID:  int(userID),
	}
	err = post1.Create()
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	post2 := &FullFeaturedPost{
		Title:   "Second Post",
		Content: "This is the second post",
		UserID:  int(userID),
	}
	err = post2.Create()
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Create profile for user
	profile := &FullFeaturedProfile{
		Bio:    "Software developer",
		UserID: int(userID),
	}
	err = profile.Create()
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Test associations
	// Load user's posts
	var posts []*FullFeaturedPost
	err = Where(&posts, "user_id = ?", user.GetID())
	if err != nil {
		t.Fatalf("Failed to load posts: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	// Load user's profile
	var profiles []*FullFeaturedProfile
	err = Where(&profiles, "user_id = ?", user.GetID())
	if err != nil {
		t.Fatalf("Failed to load profile: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}
}

func testQueryBuilder(t *testing.T) {
	// Create test data
	users := []*FullFeaturedUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, user := range users {
		err := user.Create()
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Test Query Builder
	qb := NewQueryBuilder("full_featured_users")
	qb.Where("age > ?", 25).OrderBy("age", "ASC").Limit(2)

	var youngUsers []*FullFeaturedUser
	err := qb.Find(&youngUsers)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if len(youngUsers) != 2 {
		t.Errorf("Expected 2 users, got %d", len(youngUsers))
	}

	// Test count
	count, err := qb.Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test exists
	exists, err := qb.Exists()
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}

	if !exists {
		t.Error("Expected records to exist")
	}

	// Test dry run mode
	qb.SetMode(DryRunMode)
	err = qb.Find(&youngUsers)
	if err != nil {
		t.Fatalf("Dry run should not fail: %v", err)
	}
}

func testTransactions(t *testing.T) {
	// Test simple transaction
	err := Transactional(func(tx *Transaction) error {
		user := &FullFeaturedUser{
			Name:  "Transaction User",
			Email: "tx@example.com",
			Age:   40,
		}

		// Create user within transaction
		fields, values := getFieldsAndValues(user, false)
		placeholders := make([]string, len(fields))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			user.TableName(),
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "),
		)

		result, err := tx.Exec(query, values...)
		if err != nil {
			return err
		}

		if id, err := result.LastInsertId(); err == nil {
			user.SetID(id)
		}

		// Create post within same transaction
		userID := user.GetID().(int64)
		post := &FullFeaturedPost{
			Title:   "Transaction Post",
			Content: "Created within transaction",
			UserID:  int(userID),
		}

		fields, values = getFieldsAndValues(post, false)
		for i := range placeholders {
			placeholders[i] = "?"
		}

		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			post.TableName(),
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "),
		)

		_, err = tx.Exec(query, values...)
		return err
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Test nested transactions with savepoints
	err = Transactional(func(tx *Transaction) error {
		// Create savepoint
		err := tx.CreateSavepoint("sp1")
		if err != nil {
			return err
		}

		// Do some work
		user := &FullFeaturedUser{
			Name:  "Nested User",
			Email: "nested@example.com",
			Age:   45,
		}

		fields, values := getFieldsAndValues(user, false)
		placeholders := make([]string, len(fields))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			user.TableName(),
			strings.Join(fields, ", "),
			strings.Join(placeholders, ", "),
		)

		_, err = tx.Exec(query, values...)
		if err != nil {
			return err
		}

		// Rollback to savepoint
		err = tx.RollbackToSavepoint("sp1")
		if err != nil {
			return err
		}

		// Verify rollback (user should not exist)
		rows, err := tx.Query("SELECT COUNT(*) FROM full_featured_users WHERE email = ?", "nested@example.com")
		if err != nil {
			return err
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("expected count 0 after rollback, got %d", count)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Nested transaction failed: %v", err)
	}
}

func testBatchOperations(t *testing.T) {
	// Drop and recreate table to ensure correct schema
	_, err := Exec("DROP TABLE IF EXISTS full_featured_users")
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	createTableQueries := []string{
		`CREATE TABLE IF NOT EXISTS full_featured_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER NOT NULL,
			database TEXT
		)`,
	}

	for _, query := range createTableQueries {
		_, err := Exec(query)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Direct insert test
	_, err = Exec(`INSERT INTO full_featured_users (created_at, updated_at, name, email, age) VALUES (?, ?, ?, ?, ?)`, time.Now(), time.Now(), "Direct User", "direct@example.com", 99)
	if err != nil {
		t.Fatalf("Direct insert failed: %v", err)
	}
	row := QueryRow(`SELECT COUNT(*) FROM full_featured_users`)
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("Count query failed: %v", err)
	}
	t.Logf("Direct insert count: %d", count)

	// Test batch insert
	users := []interface{}{
		&FullFeaturedUser{Name: "Batch User 1", Email: "batch1@example.com", Age: 20},
		&FullFeaturedUser{Name: "Batch User 2", Email: "batch2@example.com", Age: 21},
		&FullFeaturedUser{Name: "Batch User 3", Email: "batch3@example.com", Age: 22},
	}

	result, err := BatchInsert(users)
	if err != nil {
		t.Fatalf("Batch insert failed: %v", err)
	}
	if len(result.Errors) > 0 {
		for _, berr := range result.Errors {
			t.Logf("BatchInsert error: %v", berr)
		}
	}

	if result.RowsAffected != 3 {
		t.Errorf("Expected 3 rows affected, got %d", result.RowsAffected)
	}

	// Test find in batches
	var processedUsers []*FullFeaturedUser
	err = FindInBatches(&FullFeaturedUser{}, 2, func(batch []interface{}) error {
		for _, u := range batch {
			if user, ok := u.(*FullFeaturedUser); ok {
				processedUsers = append(processedUsers, user)
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Find in batches failed: %v", err)
	}

	if len(processedUsers) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(processedUsers))
	}

	// Test find or create
	user := &FullFeaturedUser{Name: "FindOrCreate User", Email: "findorcreate@example.com", Age: 30}
	conditions := map[string]interface{}{"email": "findorcreate@example.com"}

	err = FindOrCreate(user, conditions)
	if err != nil {
		t.Fatalf("Find or create failed: %v", err)
	}

	// Try to find or create again (should find existing)
	user2 := &FullFeaturedUser{}
	err = FindOrCreate(user2, conditions)
	if err != nil {
		t.Fatalf("Find or create failed on second attempt: %v", err)
	}

	if user2.GetID() != user.GetID() {
		t.Error("Find or create should return the same user")
	}
}

func testDatabaseResolver(t *testing.T) {
	// Use a file-based database for this test
	dbFile := "test_database_resolver.db"
	_ = os.Remove(dbFile) // Remove if exists
	defer os.Remove(dbFile)

	// Create database manager
	dm := NewDatabaseManager()

	// Create primary resolver
	primaryResolver := NewDatabaseResolver()

	// Configure primary database
	primaryConfig := &DatabaseConfig{
		Driver:   "sqlite3",
		DSN:      dbFile,
		MaxOpen:  10,
		MaxIdle:  5,
		Lifetime: time.Hour,
	}

	err := primaryResolver.SetPrimary(primaryConfig)
	if err != nil {
		t.Fatalf("Failed to set primary database: %v", err)
	}

	// Add read replica (same as primary for SQLite)
	readConfig := &DatabaseConfig{
		Driver:   "sqlite3",
		DSN:      dbFile,
		MaxOpen:  10,
		MaxIdle:  5,
		Lifetime: time.Hour,
	}

	err = primaryResolver.AddReadReplica(readConfig)
	if err != nil {
		t.Fatalf("Failed to add read replica: %v", err)
	}

	// Add write replica (same as primary for SQLite)
	writeConfig := &DatabaseConfig{
		Driver:   "sqlite3",
		DSN:      dbFile,
		MaxOpen:  10,
		MaxIdle:  5,
		Lifetime: time.Hour,
	}

	err = primaryResolver.AddWriteReplica(writeConfig)
	if err != nil {
		t.Fatalf("Failed to add write replica: %v", err)
	}

	// Add database to manager
	dm.AddDatabase("testdb", primaryResolver)

	// Set as global database manager
	SetDatabaseManager(dm)

	// Create tables on the new database
	db, err := dm.GetConnection("testdb", Primary)
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}

	// Create tables
	createTableQueries := []string{
		`CREATE TABLE IF NOT EXISTS full_featured_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER NOT NULL,
			database TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS full_featured_posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			database TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS full_featured_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bio TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			database TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
	}

	for _, query := range createTableQueries {
		_, err = db.Exec(query)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Test database-aware operations
	user := &FullFeaturedUser{
		Name:     "Database User",
		Email:    "db@example.com",
		Age:      35,
		Database: "testdb",
	}

	// Create on specific database
	err = CreateOnDatabase("testdb", user)
	if err != nil {
		t.Fatalf("Failed to create user on database: %v", err)
	}

	// Find on specific database
	foundUser := &FullFeaturedUser{}
	err = FindOnDatabase("testdb", foundUser, user.GetID())
	if err != nil {
		t.Fatalf("Failed to find user on database: %v", err)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	// Test health check
	health := dm.HealthCheck()
	if len(health) == 0 {
		t.Error("Health check should return results")
	}

	// Clean up
	err = dm.Close()
	if err != nil {
		t.Fatalf("Failed to close database manager: %v", err)
	}
}

func testLoggingAndPerformance(t *testing.T) {
	// Test structured logging
	logger := NewStructuredLogger()
	logger.SetLevel(DebugLevel)
	SetLogger(logger)

	// Create user with logging
	user := &FullFeaturedUser{
		Name:  "Log User",
		Email: "log@example.com",
		Age:   28,
	}

	// Set timestamps
	now := time.Now()
	user.SetCreatedAt(now)
	user.SetUpdatedAt(now)

	// Use logged operations (exclude ID field)
	fields, values := getFieldsAndValues(user, true)
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		user.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := LoggedExec(query, values...)
	if err != nil {
		t.Fatalf("Logged exec failed: %v", err)
	}

	if id, err := result.LastInsertId(); err == nil {
		user.SetID(id)
	}

	// Test performance metrics
	stats := GetPerformanceStats()
	if stats["total_queries"].(int64) == 0 {
		t.Error("Performance metrics should track queries")
	}

	// Test application logging
	LogInfo("Test application log", map[string]interface{}{
		"user_id": user.GetID(),
		"action":  "test",
	})

	LogDebug("Debug message", map[string]interface{}{
		"debug_info": "some debug information",
	})

	// Reset performance stats
	ResetPerformanceStats()
	stats = GetPerformanceStats()
	if stats["total_queries"].(int64) != 0 {
		t.Error("Performance stats should be reset")
	}
}

// Benchmark tests for performance
func BenchmarkBatchInsert(b *testing.B) {
	// Setup
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	createTables(&testing.T{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		users := make([]interface{}, 100)
		for j := 0; j < 100; j++ {
			users[j] = &FullFeaturedUser{
				Name:  fmt.Sprintf("User %d", j),
				Email: fmt.Sprintf("user%d@example.com", j),
				Age:   20 + j,
			}
		}

		_, err := BatchInsert(users)
		if err != nil {
			b.Fatalf("Batch insert failed: %v", err)
		}
	}
}

func BenchmarkQueryBuilder(b *testing.B) {
	// Setup
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	createTables(&testing.T{})

	// Create test data
	for i := 0; i < 1000; i++ {
		user := &FullFeaturedUser{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20 + (i % 50),
		}
		user.Create()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		qb := NewQueryBuilder("full_featured_users")
		qb.Where("age > ?", 25).OrderBy("age", "ASC").Limit(10)

		var users []*FullFeaturedUser
		err := qb.Find(&users)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}
