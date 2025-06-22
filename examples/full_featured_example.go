package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Forester-Co/go-active-record/activerecord"
)

// Example models for a blog system
type User struct {
	activerecord.HookableModel
	Name     string   `db:"name" json:"name"`
	Email    string   `db:"email" json:"email"`
	Age      int      `db:"age" json:"age"`
	Posts    []*Post  `db:"-" json:"posts"`
	Profile  *Profile `db:"-" json:"profile"`
	Database string   `db:"database" json:"database"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) SetupHooks() {
	u.AddHook(activerecord.BeforeCreate, func(m interface{}) error {
		fmt.Printf("Before creating user: %s\n", u.Name)
		return nil
	})

	u.AddHook(activerecord.AfterCreate, func(m interface{}) error {
		fmt.Printf("After creating user: %s with ID %v\n", u.Name, u.GetID())
		return nil
	})

	u.AddHook(activerecord.BeforeUpdate, func(m interface{}) error {
		fmt.Printf("Before updating user: %s\n", u.Name)
		return nil
	})
}

type Post struct {
	activerecord.HookableModel
	Title    string `db:"title" json:"title"`
	Content  string `db:"content" json:"content"`
	UserID   int    `db:"user_id" json:"user_id"`
	User     *User  `db:"-" json:"user"`
	Database string `db:"database" json:"database"`
}

func (p *Post) TableName() string {
	return "posts"
}

type Profile struct {
	activerecord.HookableModel
	Bio      string `db:"bio" json:"bio"`
	UserID   int    `db:"user_id" json:"user_id"`
	User     *User  `db:"-" json:"user"`
	Database string `db:"database" json:"database"`
}

func (p *Profile) TableName() string {
	return "profiles"
}

func main() {
	// Setup database connection
	db, err := activerecord.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup logging
	logger := activerecord.NewStructuredLogger()
	logger.SetLevel(activerecord.DebugLevel)
	activerecord.SetLogger(logger)

	// Create tables
	createTables()

	// Example 1: Basic CRUD with hooks
	exampleBasicCRUD()

	// Example 2: Associations and relationships
	exampleAssociations()

	// Example 3: Query Builder with advanced features
	exampleQueryBuilder()

	// Example 4: Transactions and savepoints
	exampleTransactions()

	// Example 5: Batch operations
	exampleBatchOperations()

	// Example 6: Database resolver with multiple databases
	exampleDatabaseResolver()

	// Example 7: Logging and performance monitoring
	exampleLoggingAndPerformance()

	// Example 8: Advanced query features
	exampleAdvancedQueries()

	fmt.Println("\n=== All examples completed successfully! ===")
}

func createTables() {
	// Create users table
	_, err := activerecord.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER,
			database TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	// Create posts table
	_, err = activerecord.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT,
			user_id INTEGER,
			database TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)
	`)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}

	// Create profiles table
	_, err = activerecord.Exec(`
		CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bio TEXT,
			user_id INTEGER UNIQUE,
			database TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)
	`)
	if err != nil {
		log.Fatal("Failed to create profiles table:", err)
	}
}

func exampleBasicCRUD() {
	fmt.Println("\n=== Example 1: Basic CRUD Operations ===")

	// Create user
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	err := user.Create()
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	fmt.Printf("Created user: %+v\n", user)

	// Find user
	var foundUser User
	err = activerecord.Find(&foundUser, user.GetID())
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return
	}

	fmt.Printf("Found user: %+v\n", foundUser)

	// Update user
	foundUser.Age = 31
	err = foundUser.Update()
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return
	}

	fmt.Printf("Updated user: %+v\n", foundUser)

	// Find all users
	var allUsers []User
	err = activerecord.FindAll(&allUsers)
	if err != nil {
		log.Printf("Failed to find all users: %v", err)
		return
	}

	fmt.Printf("All users: %+v\n", allUsers)

	// Delete user
	err = foundUser.Delete()
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return
	}

	fmt.Println("User deleted successfully")
}

func exampleAssociations() {
	fmt.Println("\n=== Example 2: Associations ===")

	// Create user with posts
	user := &User{
		Name:  "Jane Smith",
		Email: "jane@example.com",
		Age:   25,
	}

	err := user.Create()
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	// Create posts for user
	post1 := &Post{
		Title:   "First Post",
		Content: "This is my first post",
		UserID:  int(user.GetID().(int64)),
	}

	post2 := &Post{
		Title:   "Second Post",
		Content: "This is my second post",
		UserID:  int(user.GetID().(int64)),
	}

	err = post1.Create()
	if err != nil {
		log.Printf("Failed to create post1: %v", err)
		return
	}

	err = post2.Create()
	if err != nil {
		log.Printf("Failed to create post2: %v", err)
		return
	}

	// Create profile for user
	profile := &Profile{
		Bio:    "Software developer and blogger",
		UserID: int(user.GetID().(int64)),
	}

	err = profile.Create()
	if err != nil {
		log.Printf("Failed to create profile: %v", err)
		return
	}

	fmt.Printf("Created user with posts and profile: %+v\n", user)
	fmt.Printf("Posts: %+v, %+v\n", post1, post2)
	fmt.Printf("Profile: %+v\n", profile)
}

func exampleQueryBuilder() {
	fmt.Println("\n=== Example 3: Query Builder ===")

	// Create some test users
	users := []*User{
		{Name: "Alice", Email: "alice@example.com", Age: 28},
		{Name: "Bob", Email: "bob@example.com", Age: 32},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, user := range users {
		err := user.Create()
		if err != nil {
			log.Printf("Failed to create user %s: %v", user.Name, err)
		}
	}

	// Use query builder to find users
	var foundUsers []User
	err := activerecord.Where(&foundUsers, "age > ?", 30)
	if err != nil {
		log.Printf("Failed to find users: %v", err)
		return
	}

	fmt.Printf("Users older than 30: %+v\n", foundUsers)

	// Find user by email
	var user User
	err = activerecord.Where(&user, "email = ?", "alice@example.com")
	if err != nil {
		log.Printf("Failed to find user by email: %v", err)
		return
	}

	fmt.Printf("User with email alice@example.com: %+v\n", user)
}

func exampleTransactions() {
	fmt.Println("\n=== Example 4: Transactions ===")

	// Simple transaction
	err := activerecord.Transactional(func(tx *activerecord.Transaction) error {
		user := &User{
			Name:  "Transaction User",
			Email: "tx@example.com",
			Age:   40,
		}

		// Create user within transaction
		err := user.Create()
		if err != nil {
			return err
		}

		// Create post within same transaction
		post := &Post{
			Title:   "Transaction Post",
			Content: "Created within transaction",
			UserID:  int(user.GetID().(int64)),
		}

		err = post.Create()
		if err != nil {
			return err
		}

		fmt.Printf("Created user and post in transaction: %+v, %+v\n", user, post)
		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}

	fmt.Println("Transaction completed successfully")
}

func exampleBatchOperations() {
	fmt.Println("\n=== Example 5: Batch Operations ===")

	// Create multiple users
	users := []*User{
		{Name: "Batch User 1", Email: "batch1@example.com", Age: 20},
		{Name: "Batch User 2", Email: "batch2@example.com", Age: 21},
		{Name: "Batch User 3", Email: "batch3@example.com", Age: 22},
	}

	for _, user := range users {
		err := user.Create()
		if err != nil {
			log.Printf("Failed to create user %s: %v", user.Name, err)
		}
	}

	fmt.Printf("Created %d users in batch\n", len(users))

	// Find all users
	var allUsers []User
	err := activerecord.FindAll(&allUsers)
	if err != nil {
		log.Printf("Failed to find all users: %v", err)
		return
	}

	fmt.Printf("Total users in database: %d\n", len(allUsers))
}

func exampleDatabaseResolver() {
	fmt.Println("\n=== Example 6: Database Resolver ===")

	// Create a simple database resolver
	resolver := activerecord.NewDatabaseResolver()

	// Set primary database (using current connection)
	primaryConfig := &activerecord.DatabaseConfig{
		Driver:   "sqlite3",
		DSN:      ":memory:",
		MaxOpen:  10,
		MaxIdle:  5,
		Lifetime: time.Hour,
	}

	err := resolver.SetPrimary(primaryConfig)
	if err != nil {
		log.Printf("Failed to set primary database: %v", err)
		return
	}

	fmt.Println("Database resolver configured successfully")
}

func exampleLoggingAndPerformance() {
	fmt.Println("\n=== Example 7: Logging and Performance ===")

	// Create a logger
	logger := activerecord.NewDefaultLogger()
	logger.SetLevel(activerecord.InfoLevel)

	// Set global logger
	activerecord.SetLogger(logger)

	// Perform some operations with logging
	user := &User{
		Name:  "Logging User",
		Email: "logging@example.com",
		Age:   29,
	}

	err := user.Create()
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	fmt.Printf("Created user with logging: %+v\n", user)
}

func exampleAdvancedQueries() {
	fmt.Println("\n=== Example 8: Advanced Queries ===")

	// Create test data
	user := &User{
		Name:  "Advanced User",
		Email: "advanced@example.com",
		Age:   33,
	}

	err := user.Create()
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	// Create posts for the user
	posts := []*Post{
		{Title: "Advanced Post 1", Content: "Content 1", UserID: int(user.GetID().(int64))},
		{Title: "Advanced Post 2", Content: "Content 2", UserID: int(user.GetID().(int64))},
	}

	for _, post := range posts {
		err := post.Create()
		if err != nil {
			log.Printf("Failed to create post: %v", err)
		}
	}

	// Find user with posts
	var foundUser User
	err = activerecord.Find(&foundUser, user.GetID())
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return
	}

	// Find posts for user
	var userPosts []Post
	err = activerecord.Where(&userPosts, "user_id = ?", user.GetID())
	if err != nil {
		log.Printf("Failed to find posts: %v", err)
		return
	}

	fmt.Printf("User: %+v\n", foundUser)
	fmt.Printf("User's posts: %+v\n", userPosts)
}

// Helper function to join strings (since strings.Join is not available in this context)
func Join(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}
