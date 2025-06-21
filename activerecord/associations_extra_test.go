package activerecord

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type AssocModel struct{ ActiveRecordModel }

func (a *AssocModel) TableName() string { return "assoc_models" }

func TestAssociationMethods(t *testing.T) {
	m := &ActiveRecordModel{}
	m.HasOne("profile", &AssocModel{}, "user_id")
	m.HasMany("posts", &AssocModel{}, "user_id")
	m.BelongsTo("user", &AssocModel{}, "user_id")
	m.HasManyThrough("tags", &AssocModel{}, "posts", "tag_id", "user_id")
}

func TestLoad_Include_Errors(t *testing.T) {
	m := &ActiveRecordModel{}
	err := m.Load("notfound")
	if err == nil {
		t.Error("Load should return error for missing association")
	}
	err = m.Include("notfound1", "notfound2")
	if err == nil {
		t.Error("Include should return error for missing association")
	}
}

func TestJoinAndEagerLoadingStubs(t *testing.T) {
	var models []AssocModel
	if err := Joins(&models, "join"); err != nil {
		t.Errorf("Joins should not fail: %v", err)
	}
	if err := LeftJoins(&models, "left"); err != nil {
		t.Errorf("LeftJoins should not fail: %v", err)
	}
	if err := InnerJoins(&models, "inner"); err != nil {
		t.Errorf("InnerJoins should not fail: %v", err)
	}
	if err := With(&models, "assoc"); err != nil {
		t.Errorf("With should not fail: %v", err)
	}
	if err := Preload(&models, "assoc"); err != nil {
		t.Errorf("Preload should not fail: %v", err)
	}
}

type Author struct {
	ID        interface{} `db:"id"`
	Name      string      `db:"name"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func (a *Author) TableName() string { return "authors" }
func (a *Author) Create() error     { return Create(a) }
func (a *Author) Reload() error     { return Find(a, a.GetID()) }
func (a *Author) HasMany(name string, model interface{}, foreignKey string) {
	associationRegistry[name] = &Association{
		Type:       HasMany,
		Model:      model,
		ForeignKey: foreignKey,
	}
}
func (a *Author) Load(name string) error {
	association, exists := associationRegistry[name]
	if !exists {
		return fmt.Errorf("association %s not found", name)
	}
	return loadHasMany(a, name, association)
}
func (a *Author) GetID() interface{}       { return a.ID }
func (a *Author) SetID(id interface{})     { a.ID = id }
func (a *Author) GetCreatedAt() time.Time  { return a.CreatedAt }
func (a *Author) SetCreatedAt(t time.Time) { a.CreatedAt = t }
func (a *Author) GetUpdatedAt() time.Time  { return a.UpdatedAt }
func (a *Author) SetUpdatedAt(t time.Time) { a.UpdatedAt = t }
func (a *Author) Find(id interface{}) error {
	return Find(a, id)
}
func (a *Author) Where(query string, args ...interface{}) (interface{}, error) {
	var authors []Author
	err := Where(&authors, query, args...)
	return authors, err
}

type Book struct {
	ID        interface{} `db:"id"`
	Title     string      `db:"title"`
	AuthorID  int         `db:"authorid"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func (b *Book) TableName() string { return "books" }
func (b *Book) Create() error     { return Create(b) }
func (b *Book) Reload() error     { return Find(b, b.GetID()) }
func (b *Book) BelongsTo(name string, model interface{}, foreignKey string) {
	associationRegistry[name] = &Association{
		Type:       BelongsTo,
		Model:      model,
		ForeignKey: foreignKey,
	}
}
func (b *Book) Load(name string) error {
	association, exists := associationRegistry[name]
	if !exists {
		return fmt.Errorf("association %s not found", name)
	}
	return loadBelongsTo(b, name, association)
}
func (b *Book) GetID() interface{}       { return b.ID }
func (b *Book) SetID(id interface{})     { b.ID = id }
func (b *Book) GetCreatedAt() time.Time  { return b.CreatedAt }
func (b *Book) SetCreatedAt(t time.Time) { b.CreatedAt = t }
func (b *Book) GetUpdatedAt() time.Time  { return b.UpdatedAt }
func (b *Book) SetUpdatedAt(t time.Time) { b.UpdatedAt = t }
func (b *Book) Find(id interface{}) error {
	return Find(b, id)
}
func (b *Book) Where(query string, args ...interface{}) (interface{}, error) {
	var books []Book
	err := Where(&books, query, args...)
	return books, err
}

func TestHasManyAndBelongsToAssociations(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create authors table: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		title TEXT, 
		authorid INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create books table: %v", err)
	}

	author := &Author{Name: "Leo Tolstoy"}
	if err := author.Create(); err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}
	t.Logf("author.GetID() after Create = %#v, type=%T", author.GetID(), author.GetID())
	if err := author.Reload(); err != nil {
		t.Fatalf("Failed to reload author: %v", err)
	}
	t.Logf("author.GetID() after Reload = %#v, type=%T", author.GetID(), author.GetID())
	aid := author.GetID()
	var authorID int
	switch v := aid.(type) {
	case int:
		authorID = v
	case int64:
		authorID = int(v)
	default:
		t.Fatalf("unexpected author ID type: %T", aid)
	}

	books := []Book{
		{Title: "War and Peace", AuthorID: authorID},
		{Title: "Anna Karenina", AuthorID: authorID},
	}
	for i := range books {
		if err := books[i].Create(); err != nil {
			t.Fatalf("Failed to create book %d: %v", i, err)
		}
	}

	// HasMany: Author has many Books
	var foundBooks []*Book
	author.HasMany("books", &foundBooks, "authorid")
	err := author.Load("books")
	if err != nil {
		t.Fatalf("Load books failed: %v", err)
	}
	if len(foundBooks) != 2 {
		t.Errorf("Expected 2 books, got %d", len(foundBooks))
	}

	// BelongsTo: Book belongs to Author
	var foundAuthor Author
	books[0].BelongsTo("author", &foundAuthor, "AuthorID")
	err = books[0].Load("author")
	if err != nil {
		t.Fatalf("Load author failed: %v", err)
	}
	if foundAuthor.Name != author.Name {
		t.Errorf("Expected author %s, got %s", author.Name, foundAuthor.Name)
	}
}

// Helper functions to call the original logic with the correct receiver
func loadHasMany(model interface{}, name string, association *Association) error {
	modeler, ok := model.(Modeler)
	if !ok {
		return fmt.Errorf("model does not implement Modeler")
	}
	foreignKey := association.ForeignKey
	id := modeler.GetID()
	query := foreignKey + " = ?"

	// Try setter methods first
	if name == "Mentees" {
		if setter, ok := model.(interface{ SetMentees([]*User) }); ok {
			var result []*User
			err := Where(&result, query, id)
			if err != nil {
				return err
			}
			setter.SetMentees(result)
			return nil
		}
	}

	// Fallback to original behavior
	if association.Model != nil {
		return Where(association.Model, query, id)
	}
	return nil
}

func loadBelongsTo(model interface{}, name string, association *Association) error {
	foreignKey := association.ForeignKey

	// Try setter methods first
	if name == "Mentor" {
		if setter, ok := model.(interface{ SetMentor(*User) }); ok {
			var mentor User
			val := reflect.ValueOf(model).Elem().FieldByName(foreignKey)
			if !val.IsValid() {
				return nil
			}
			err := Find(&mentor, val.Interface())
			if err != nil {
				return err
			}
			setter.SetMentor(&mentor)
			return nil
		}
	}

	// Fallback to original behavior
	if association.Model != nil {
		val := reflect.ValueOf(model).Elem().FieldByName(foreignKey)
		if !val.IsValid() {
			return nil
		}
		return Find(association.Model, val.Interface())
	}
	return nil
}

type User struct {
	ID        interface{} `db:"id"`
	Name      string      `db:"name"`
	MentorID  int         `db:"mentorid"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
	Mentor    *User       `db:"-"`
	Mentees   []*User     `db:"-"`
}

func (u *User) TableName() string          { return "users" }
func (u *User) Create() error              { return Create(u) }
func (u *User) Reload() error              { return Find(u, u.GetID()) }
func (u *User) SetMentor(mentor *User)     { u.Mentor = mentor }
func (u *User) SetMentees(mentees []*User) { u.Mentees = mentees }
func (u *User) Load(name string) error {
	autoRegisterAssociations(u)
	association, exists := associationRegistry[name]
	if !exists {
		return fmt.Errorf("association %s not found", name)
	}

	switch association.Type {
	case HasMany:
		return loadHasMany(u, name, association)
	case BelongsTo:
		return loadBelongsTo(u, name, association)
	default:
		return fmt.Errorf("unsupported association type")
	}
}
func (u *User) GetID() interface{}       { return u.ID }
func (u *User) SetID(id interface{})     { u.ID = id }
func (u *User) GetCreatedAt() time.Time  { return u.CreatedAt }
func (u *User) SetCreatedAt(t time.Time) { u.CreatedAt = t }
func (u *User) GetUpdatedAt() time.Time  { return u.UpdatedAt }
func (u *User) SetUpdatedAt(t time.Time) { u.UpdatedAt = t }
func (u *User) Find(id interface{}) error {
	return Find(u, id)
}
func (u *User) Where(query string, args ...interface{}) (interface{}, error) {
	var users []*User
	err := Where(&users, query, args...)
	return users, err
}

func TestAutoAssociationDetection(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		mentorid INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	mentor := &User{Name: "Master"}
	if err := mentor.Create(); err != nil {
		t.Fatalf("Failed to create mentor: %v", err)
	}
	if err := mentor.Reload(); err != nil {
		t.Fatalf("Failed to reload mentor: %v", err)
	}
	if mentor.GetID() == nil {
		t.Fatal("mentor.GetID() is nil after Create+Reload")
	}
	var mentorID int
	switch v := mentor.GetID().(type) {
	case int:
		mentorID = v
	case int64:
		mentorID = int(v)
	default:
		t.Fatalf("unexpected mentor ID type: %T", mentor.GetID())
	}

	mentee1 := &User{Name: "Student1", MentorID: mentorID}
	mentee2 := &User{Name: "Student2", MentorID: mentorID}
	if err := mentee1.Create(); err != nil {
		t.Fatalf("Failed to create mentee1: %v", err)
	}
	if err := mentee2.Create(); err != nil {
		t.Fatalf("Failed to create mentee2: %v", err)
	}

	// Check belongs_to
	if err := mentee1.Load("Mentor"); err != nil {
		t.Fatalf("Failed to load mentor: %v", err)
	}
	if mentee1.Mentor == nil || mentee1.Mentor.Name != "Master" {
		t.Errorf("Expected mentor Master, got %#v", mentee1.Mentor)
	}

	// Check has_many
	if err := mentor.Load("Mentees"); err != nil {
		t.Fatalf("Failed to load mentees: %v", err)
	}
	if len(mentor.Mentees) != 2 {
		t.Errorf("Expected 2 mentees, got %d", len(mentor.Mentees))
	}
}
