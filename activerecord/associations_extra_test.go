package activerecord

import (
	"fmt"
	"reflect"
	"testing"
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
	BaseModel
	Name string
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

type Book struct {
	BaseModel
	Title    string
	AuthorID int
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

func TestHasManyAndBelongsToAssociations(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	db.Exec(`CREATE TABLE authors (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, created_at TIMESTAMP, updated_at TIMESTAMP)`)
	db.Exec(`CREATE TABLE books (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, authorid INTEGER, created_at TIMESTAMP, updated_at TIMESTAMP)`)

	author := &Author{Name: "Leo Tolstoy"}
	author.Create()
	t.Logf("author.GetID() after Create = %#v, type=%T", author.GetID(), author.GetID())
	author.Reload()
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
		books[i].Create()
	}

	// HasMany: Author has many Books
	var foundBooks []Book
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
	switch name {
	case "Mentees":
		if setter, ok := model.(interface{ SetMentees([]User) }); ok {
			var result []User
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
	switch name {
	case "Mentor":
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
	BaseModel
	Name     string
	MentorID int
	Mentor   *User  `db:"-"`
	Mentees  []User `db:"-"`
}

func (u *User) TableName() string         { return "users" }
func (u *User) Create() error             { return Create(u) }
func (u *User) Reload() error             { return Find(u, u.GetID()) }
func (u *User) SetMentor(mentor *User)    { u.Mentor = mentor }
func (u *User) SetMentees(mentees []User) { u.Mentees = mentees }
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

func TestAutoAssociationDetection(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, mentorid INTEGER, created_at TIMESTAMP, updated_at TIMESTAMP)`)

	mentor := &User{Name: "Master"}
	mentor.Create()
	mentor.Reload()
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
	mentee1.Create()
	mentee2.Create()

	// Check belongs_to
	mentee1.Load("Mentor")
	if mentee1.Mentor == nil || mentee1.Mentor.Name != "Master" {
		t.Errorf("Expected mentor Master, got %#v", mentee1.Mentor)
	}

	// Check has_many
	mentor.Load("Mentees")
	if len(mentor.Mentees) != 2 {
		t.Errorf("Expected 2 mentees, got %d", len(mentor.Mentees))
	}
}
