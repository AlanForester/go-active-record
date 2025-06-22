# Go Active Record

A full-featured Active Record ORM for Go, inspired by Rails Active Record. Provides a comprehensive interface for database operations with advanced features like hooks, batch operations, database resolver, query builder, and more.

## 🚀 Features

### Core Features
- **CRUD Operations** - Create, Read, Update, Delete records with automatic ID management
- **Hooks System** - Before/After hooks for Create, Update, Delete, Save, Find operations
- **Transaction Support** - Full transaction management with context support
- **Query Builder** - Fluent interface for building complex queries with dry run mode
- **Batch Operations** - Efficient batch insert, find in batches, and bulk operations
- **Database Resolver** - Multi-database support with primary/read/write replica management
- **Logging & Performance** - Structured logging and performance metrics tracking

### Data Validation
- **Validations** - Built-in validators for data validation (presence, length, email, numericality, format)
- **Error Collection** - Comprehensive error collection and reporting

### Database Management
- **Migrations** - Database schema management with version control
- **Table Builder** - DSL for creating and modifying tables
- **Connection Pooling** - Efficient database connection management
- **Health Checks** - Database health monitoring

### Advanced Features
- **NULL Value Handling** - Proper handling of NULL database values
- **Reflection-based Mapping** - Dynamic field discovery for complex structs
- **Association Framework** - Relationship management between models
- **Context Support** - Context-aware operations for cancellation and timeouts

## 📦 Installation

```bash
go get github.com/Forester-Co/go-active-record
```

## 🚀 Quick Start

### Database Connection

```go
package main

import (
    "log"
    "github.com/Forester-Co/go-active-record/activerecord"
)

func main() {
    // Connect to SQLite (for development)
    db, err := activerecord.Connect("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Set global connection
    activerecord.SetConnection(db)
}
```

### Model Definition

```go
type User struct {
    activerecord.HookableModel  // Includes hooks and Active Record methods
    Name     string `db:"name" json:"name"`
    Email    string `db:"email" json:"email"`
    Age      int    `db:"age" json:"age"`
    Database string `db:"database" json:"database"`
}

// TableName returns the table name
func (u *User) TableName() string {
    return "users"
}

// SetupHooks configures model hooks
func (u *User) SetupHooks() {
    u.AddHook(activerecord.BeforeCreate, func(model interface{}) error {
        user := model.(*User)
        fmt.Printf("Creating user: %s\n", user.Name)
        return nil
    })
    
    u.AddHook(activerecord.AfterCreate, func(model interface{}) error {
        user := model.(*User)
        fmt.Printf("Created user with ID: %v\n", user.GetID())
        return nil
    })
}
```

### CRUD Operations

```go
// Create with hooks
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}
user.SetupHooks()
err := user.Create()

// Read by ID
foundUser := &User{}
err = activerecord.Find(foundUser, 1)

// Read all records
var users []*User
err = activerecord.FindAll(&users)

// Search with conditions
var youngUsers []*User
err = activerecord.Where(&youngUsers, "age < ?", 25)

// Update
foundUser.Age = 31
err = foundUser.Update()

// Delete
err = foundUser.Delete()

// Save (creates or updates)
err = user.Save()
```

### Query Builder

```go
// Create query builder
qb := activerecord.NewQueryBuilder("users")
qb.Where("age > ?", 25).
   Where("email LIKE ?", "%@example.com").
   OrderBy("age", "ASC").
   Limit(10).
   Offset(0)

// Execute query
var users []*User
err := qb.Find(&users)

// Dry run for debugging
qb.DryRun(true)
err = qb.Find(&users) // Prints the query without executing
```

### Batch Operations

```go
// Batch insert
users := []interface{}{
    &User{Name: "User 1", Email: "user1@example.com", Age: 25},
    &User{Name: "User 2", Email: "user2@example.com", Age: 30},
    &User{Name: "User 3", Email: "user3@example.com", Age: 35},
}

result, err := activerecord.BatchInsert(users)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Inserted %d users\n", result.RowsAffected)

// Find in batches
err = activerecord.FindInBatches(&User{}, 100, func(batch []interface{}) error {
    for _, user := range batch {
        // Process each user
        fmt.Printf("Processing user: %v\n", user.(*User).Name)
    }
    return nil
})

// Find or create
user := &User{Email: "new@example.com"}
conditions := map[string]interface{}{"email": "new@example.com"}
err = activerecord.FindOrCreate(user, conditions)
```

### Transactions

```go
// Begin transaction
tx, err := activerecord.Begin()
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()

// Create user within transaction
user := &User{Name: "Transaction User", Email: "tx@example.com"}
err = user.Create()
if err != nil {
    return err
}

// Create related record
profile := &Profile{UserID: user.GetID(), Bio: "Transaction test"}
err = profile.Create()
if err != nil {
    return err
}

// Commit transaction
err = tx.Commit()
```

### Database Resolver (Multi-Database Support)

```go
// Create database manager
dm := activerecord.NewDatabaseManager()

// Configure primary database
primaryResolver := activerecord.NewDatabaseResolver()
primaryConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "primary.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.SetPrimary(primaryConfig)

// Add read replica
readConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "read_replica.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.AddReadReplica(readConfig)

// Add write replica
writeConfig := &activerecord.DatabaseConfig{
    Driver:   "sqlite3",
    DSN:      "write_replica.db",
    MaxOpen:  10,
    MaxIdle:  5,
    Lifetime: time.Hour,
}
primaryResolver.AddWriteReplica(writeConfig)

// Add to manager
dm.AddDatabase("myapp", primaryResolver)
activerecord.SetDatabaseManager(dm)

// Use database-aware operations
user := &User{Name: "Multi-DB User", Email: "multidb@example.com"}
err := activerecord.CreateOnDatabase("myapp", user)

foundUser := &User{}
err = activerecord.FindOnDatabase("myapp", foundUser, user.GetID())
```

### Logging and Performance

```go
// Setup structured logging
logger := activerecord.NewStructuredLogger()
logger.SetLevel(activerecord.DebugLevel)
activerecord.SetLogger(logger)

// Logged operations
result, err := activerecord.LoggedExec("INSERT INTO users (name, email) VALUES (?, ?)", "Log User", "log@example.com")

// Performance metrics
stats := activerecord.GetPerformanceStats()
fmt.Printf("Total queries: %d\n", stats["total_queries"])

// Application logging
activerecord.LogInfo("User created", map[string]interface{}{
    "user_id": user.GetID(),
    "action":  "create",
})
```

### Validations

```go
type User struct {
    activerecord.ValidationModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
    Age   int    `db:"age" json:"age"`
}

func (u *User) TableName() string {
    return "users"
}

// Setup validations
func (u *User) SetupValidations() {
    u.PresenceOf("Name")
    u.AddValidation("Email", "email", "has invalid format")
    u.Length("Name", 2, 50)
    u.Numericality("Age", 18, 100)
    u.Format("Email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "invalid email format")
}

// Validate
user := &User{Name: "", Email: "invalid-email", Age: 15}
user.SetupValidations()

if !user.IsValid() {
    fmt.Println("Validation errors:", user.Errors())
}
```

### Migrations

```go
// Create migration
migration := activerecord.NewMigration("create_users_table")
migration.Up = func() error {
    return activerecord.CreateTable("users", func(t *activerecord.TableBuilder) {
        t.Integer("id").PrimaryKey().AutoIncrement()
        t.String("name").NotNull()
        t.String("email").NotNull().Unique()
        t.Integer("age")
        t.Timestamp("created_at").NotNull()
        t.Timestamp("updated_at").NotNull()
    })
}

migration.Down = func() error {
    return activerecord.DropTable("users")
}

// Run migration
err := migration.Migrate()

// Check migration status
status := migration.Status()
fmt.Printf("Migration status: %s\n", status)
```

## 🧪 Testing

The library includes comprehensive tests covering all features:

```bash
# Run all tests
go test ./activerecord -v

# Run specific test
go test ./activerecord -v -run TestFullFeaturedORM

# Run benchmarks
go test ./activerecord -bench=.
```

## 📊 Performance

The library is optimized for performance with features like:
- Connection pooling
- Prepared statements
- Batch operations
- Efficient reflection usage
- Memory-conscious design

## 🔧 Configuration

### Database Configuration

```go
// SQLite
db, err := activerecord.Connect("sqlite3", ":memory:")

// PostgreSQL
db, err := activerecord.Connect("postgres", "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")

// MySQL
db, err := activerecord.Connect("mysql", "user:password@tcp(localhost:3306)/testdb")
```

### Logging Configuration

```go
// Structured logging
logger := activerecord.NewStructuredLogger()
logger.SetLevel(activerecord.DebugLevel)
activerecord.SetLogger(logger)

// Custom logger
activerecord.SetLogger(customLogger)
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by Ruby on Rails Active Record
- Built with Go's standard `database/sql` package
- Uses reflection for dynamic field mapping
- Implements modern Go patterns and best practices 