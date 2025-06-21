# Go Active Record

An Active Record library for Go, inspired by Rails Active Record. Provides a convenient interface for database operations, including CRUD operations, validations, associations, and migrations.

## Features

- 🚀 **CRUD Operations** - create, read, update, delete records
- ✅ **Validations** - built-in validators for data validation
- 🔗 **Associations** - relationships between models (has_one, has_many, belongs_to)
- 📊 **Migrations** - database schema management
- 🔍 **Query Builder** - convenient query builder
- 🛡️ **Transactions** - transaction support
- 📝 **Logging** - built-in SQL query logging

## Installation

```bash
go get github.com/your-username/go-active-record
```

## Quick Start

### Database Connection

```go
package main

import (
    "log"
    "github.com/your-username/go-active-record/activerecord"
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

### Associations

The library supports automatic association detection and manual association definition.

#### Auto-Association Detection

You can define associations by simply adding fields to your struct:

```go
type User struct {
    activerecord.BaseModel
    Name     string
    MentorID int
    Mentor   *User  `db:"-"`  // BelongsTo association
    Mentees  []User `db:"-"`  // HasMany association
}

// The library automatically detects and registers associations:
// - Mentor field (*User) -> BelongsTo association with foreign key "MentorID"
// - Mentees field ([]User) -> HasMany association with foreign key "MentorID"

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
- **HasMany**: `[]OtherModel` - one-to-many relationship where this model has many others
- **HasOne**: `*OtherModel` - one-to-one relationship where this model has one other
- **HasManyThrough**: complex many-to-many relationships (planned)

#### Association Loading

```go
// Load single association
user.Load("Mentor")

// Load multiple associations
user.Include("Mentor", "Mentees")
```

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
- `IsPersisted() bool` - checks if record is persisted
- `Touch() error` - updates timestamps
- `Reload() error` - reloads data from database

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
- `Email(field string)` - check email format
- `Uniqueness(field string)` - check uniqueness
- `Numericality(field string, min, max float64)` - check numeric value
- `Format(field string, pattern string)` - check regex pattern

### Migrations

- `CreateTable(tableName string, callback func(*TableBuilder)) error` - create table
- `DropTable(tableName string) error` - drop table
- `Column(name, dataType string, options ...string)` - add column
- `PrimaryKey(columns ...string)` - add primary key
- `Index(columns ...string)` - add index
- `Timestamps()` - add timestamps

## Examples

Complete usage examples can be found in the `main.go` file.

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL/MySQL/SQLite

### Setup

```bash
# Clone the repository
git clone https://github.com/your-username/go-active-record.git
cd go-active-record

# Install dependencies
make deps

# Run tests
make test

# Build the project
make build
```

### Available Commands

```bash
make help          # Show available commands
make test          # Run tests
make build         # Build project
make lint          # Run linter
make fmt           # Format code
make clean         # Clean build artifacts
make example       # Run example
make migrate       # Run migrations
```

### Docker Development

```bash
# Start all services
docker-compose up -d

# Run tests in container
docker-compose exec app make test

# Stop all services
docker-compose down
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## TODO

- [x] Implement associations
- [ ] Add transaction support
- [ ] Add SQL query logging
- [ ] Add caching
- [ ] Support for other databases
- [ ] Add more tests
- [ ] Complete API documentation 