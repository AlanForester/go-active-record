# Go Active Record

An Active Record library for Go, inspired by Rails Active Record. Provides a convenient interface for database operations, including CRUD operations, validations, associations, and migrations.

## Features

- 🚀 **CRUD Operations** - create, read, update, delete records
- ✅ **Validations** - built-in validators for data validation
- 🔗 **Associations** - relationships between models (has_one, has_many, belongs_to)
- 🤖 **Auto-Association Detection** - automatically detect and register associations
- 📊 **Migrations** - database schema management
- 🔍 **Query Builder** - convenient query builder
- 🛡️ **Transactions** - transaction support
- 📝 **Logging** - built-in SQL query logging
- 🔧 **CI/CD** - GitHub Actions for automated testing and deployment
- 🛡️ **Security** - automated security scanning and vulnerability checks

## Installation

```bash
go get github.com/Forester-Co/go-active-record
```

## Quick Start

### Database Connection

```go
package main

import (
    "log"
    "github.com/Forester-Co/go-active-record/activerecord"
)

func main() {
    // Connect to PostgreSQL
    db, err := activerecord.Connect("postgres", "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")
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
    activerecord.ActiveRecordModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
    Age   int    `db:"age" json:"age"`
}

// TableName returns the table name
func (u *User) TableName() string {
    return "users"
}
```

### CRUD Operations

```go
// Create
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}
err := user.Create()

// Read by ID
foundUser := &User{}
err = activerecord.Find(foundUser, 1)

// Read all records
var users []User
err = activerecord.FindAll(&users)

// Search with conditions
var youngUsers []User
err = activerecord.Where(&youngUsers, "age < ?", 25)

// Update
foundUser.Age = 31
err = foundUser.Update()

// Delete
err = foundUser.Delete()
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
}

// Validate
user := &User{Name: "", Email: "invalid-email", Age: 15}
user.SetupValidations()

if !user.IsValid() {
    fmt.Println("Validation errors:", user.Errors())
}
```

### Associations

The library supports both automatic association detection and manual association definition.

#### Auto-Association Detection

You can define associations by simply adding fields to your struct:

```go
type User struct {
    activerecord.BaseModel
    Name     string
    MentorID int
    Mentor   *User  `db:"-"`  // BelongsTo association
    Mentees  []*User `db:"-"`  // HasMany association
}

// The library automatically detects and registers associations:
// - Mentor field (*User) -> BelongsTo association with foreign key "MentorID"
// - Mentees field ([]*User) -> HasMany association with foreign key "MentorID"

// Usage
mentor := &User{Name: "Master"}
mentor.Create()

mentee := &User{Name: "Student", MentorID: mentor.GetID()}
mentee.Create()

// Load associations
mentee.Load("Mentor")    // Loads the mentor
mentor.Load("Mentees")   // Loads all mentees
```

#### Manual Association Definition

You can also define associations manually:

```go
type User struct {
    activerecord.ActiveRecordModel
    Name  string `db:"name" json:"name"`
    Email string `db:"email" json:"email"`
}

type Post struct {
    activerecord.ActiveRecordModel
    Title   string `db:"title" json:"title"`
    Content string `db:"content" json:"content"`
    UserID  int    `db:"user_id" json:"user_id"`
}

// Define associations manually
func (u *User) HasMany(name string, model interface{}, foreignKey string) {
    // has_many implementation
}

func (p *Post) BelongsTo(name string, model interface{}, foreignKey string) {
    // belongs_to implementation
}
```

#### Supported Association Types

- **BelongsTo**: `*OtherModel` - one-to-one relationship where this model belongs to another
- **HasMany**: `[]OtherModel` or `[]*OtherModel` - one-to-many relationship where this model has many others
- **HasOne**: `*OtherModel` - one-to-one relationship where this model has one other
- **HasManyThrough**: complex many-to-many relationships (planned)

#### Association Loading

```go
// Load single association
user.Load("Mentor")

// Load multiple associations
user.Include("Mentor", "Mentees")
```

### Migrations

```go
type CreateUsersTable struct {
    activerecord.Migration
}

func (m *CreateUsersTable) Version() int64 {
    return 20231201000001
}

func (m *CreateUsersTable) Up() error {
    return activerecord.CreateTable("users", func(t *activerecord.TableBuilder) {
        t.Column("id", "SERIAL", "PRIMARY KEY")
        t.Column("name", "VARCHAR(255)", "NOT NULL")
        t.Column("email", "VARCHAR(255)", "UNIQUE", "NOT NULL")
        t.Column("age", "INTEGER")
        t.Timestamps()
        t.Index("email")
    })
}

func (m *CreateUsersTable) Down() error {
    return activerecord.DropTable("users")
}

// Run migrations
func main() {
    migrator := activerecord.NewMigrator()
    migrations := []activerecord.Migration{
        &CreateUsersTable{},
    }
    
    err := migrator.Migrate(migrations)
    if err != nil {
        log.Fatal(err)
    }
}
```

## CI/CD & Automation

This project includes comprehensive GitHub Actions workflows:

- **CI/CD Pipeline** - Automated testing on multiple Go versions (1.21-1.24)
- **Security Scanning** - Weekly security checks with gosec and govulncheck
- **Code Quality** - Automated linting with golangci-lint
- **Documentation** - Auto-generated docs on GitHub Pages
- **Dependency Updates** - Automated dependency updates with Dependabot
- **Release Management** - Automated releases when tags are created

## Supported Databases

- PostgreSQL
- MySQL
- SQLite

## API Reference

### Core Methods

#### ActiveRecordModel

- `Create() error` - creates a record
- `Update() error` - updates a record
- `Delete() error` - deletes a record
- `Save() error` - saves a record (creates or updates)
- `IsNewRecord() bool` - checks if record is new
- `IsPersisted() bool` - checks if record is saved
- `Touch() error` - updates timestamps
- `Reload() error` - reloads data from DB

#### Global Methods

- `Find(model Modeler, id interface{}) error` - find by ID
- `FindAll(models interface{}) error` - find all records
- `Where(models interface{}, query string, args ...interface{}) error` - find with conditions
- `Create(model Modeler) error` - create record
- `Update(model Modeler) error` - update record
- `Delete(model Modeler) error` - delete record

### Validators

- `PresenceOf(field string)` - check for presence
- `Length(field string, min, max int)` - check string length
- `Email(field string)` - validate email format
- `Uniqueness(field string)` - check uniqueness
- `Numericality(field string, min, max float64)` - validate numeric value
- `Format(field string, pattern string)` - validate with regex

### Migrations

- `CreateTable(tableName string, callback func(*TableBuilder)) error` - create table
- `DropTable(tableName string) error` - drop table
- `Column(name, dataType string, options ...string)` - add column
- `PrimaryKey(columns ...string)` - add primary key
- `Index(columns ...string)` - add index
- `Timestamps()` - add timestamps

## Examples

Complete usage examples can be found in the `examples/` directory.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Security

Please report security vulnerabilities to security@forester.co. See [SECURITY.md](SECURITY.md) for more information.

## License

MIT License

## Status

- [x] CRUD operations
- [x] Validations
- [x] Associations (manual and automatic)
- [x] Migrations
- [x] Query builder
- [x] CI/CD pipeline
- [x] Security scanning
- [ ] Transactions
- [ ] HasManyThrough associations
- [ ] Advanced query builder
- [ ] Connection pooling 