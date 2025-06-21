package activerecord

import (
	"reflect"
	"testing"
	"time"
)

type ModelTest struct {
	ID        interface{} `db:"id"`
	Name      string      `db:"name"`
	Age       int         `db:"age"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func (m *ModelTest) TableName() string        { return "model_tests" }
func (m *ModelTest) GetID() interface{}       { return m.ID }
func (m *ModelTest) SetID(id interface{})     { m.ID = id }
func (m *ModelTest) GetCreatedAt() time.Time  { return m.CreatedAt }
func (m *ModelTest) SetCreatedAt(t time.Time) { m.CreatedAt = t }
func (m *ModelTest) GetUpdatedAt() time.Time  { return m.UpdatedAt }
func (m *ModelTest) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func (m *ModelTest) Find(id interface{}) error {
	return Find(m, id)
}

func (m *ModelTest) Where(query string, args ...interface{}) (interface{}, error) {
	var results []*ModelTest
	err := Where(&results, query, args...)
	return results, err
}

func TestBaseModelMethods(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE model_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		age INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	model := &BaseModel{}
	model.SetID(1)
	if model.GetID() != 1 {
		t.Error("GetID should return 1")
	}

	now := time.Now()
	model.SetCreatedAt(now)
	if !model.GetCreatedAt().Equal(now) {
		t.Error("GetCreatedAt should return the set time")
	}

	model.SetUpdatedAt(now)
	if !model.GetUpdatedAt().Equal(now) {
		t.Error("GetUpdatedAt should return the set time")
	}
}

func TestSetTimestampsDeep(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE model_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		age INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	model := &ModelTest{Name: "Test", Age: 25}
	now := time.Now()
	setTimestampsDeep(reflect.ValueOf(model), now)

	if model.GetCreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}
	if model.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestGetFieldsAndValues(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE model_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		age INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if err := Create(&ModelTest{Name: "A", Age: 1}); err != nil {
		t.Fatalf("Failed to create model A: %v", err)
	}
	if err := Create(&ModelTest{Name: "B", Age: 2}); err != nil {
		t.Fatalf("Failed to create model B: %v", err)
	}

	m := &ModelTest{Name: "A", Age: 1}
	fields, values := getFieldsAndValues(m, false)
	if len(fields) == 0 {
		t.Error("Fields should not be empty")
	}
	if len(values) == 0 {
		t.Error("Values should not be empty")
	}
}

func TestCreate_Find_Update_Delete(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE model_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		age INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	m := &ModelTest{Name: "Test", Age: 10}
	if err := Create(m); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	m.Age = 20
	if err := Update(m); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	found := &ModelTest{}
	if err := Find(found, m.GetID()); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if err := Delete(m); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestFindAll_Where(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE model_tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT, 
		age INTEGER, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	if err := Create(&ModelTest{Name: "A", Age: 1}); err != nil {
		t.Fatalf("Failed to create model A: %v", err)
	}
	if err := Create(&ModelTest{Name: "B", Age: 2}); err != nil {
		t.Fatalf("Failed to create model B: %v", err)
	}
	var all []*ModelTest
	if err := FindAll(&all); err != nil || len(all) < 2 {
		t.Fatalf("FindAll failed or not enough records: %v", err)
	}
	var filtered []*ModelTest
	if err := Where(&filtered, "age = ?", 2); err != nil || len(filtered) != 1 {
		t.Fatalf("Where failed or wrong number of records: %v", err)
	}
}

func TestCreate_ModelerError(t *testing.T) {
	err := Create(struct{}{})
	if err == nil {
		t.Error("Create should fail for non-Modeler")
	}
}

func TestFind_ModelerError(t *testing.T) {
	err := Find(struct{}{}, 1)
	if err == nil {
		t.Error("Find should fail for non-Modeler")
	}
}

func TestUpdate_ModelerError(t *testing.T) {
	err := Update(struct{}{})
	if err == nil {
		t.Error("Update should fail for non-Modeler")
	}
}

func TestDelete_ModelerError(t *testing.T) {
	err := Delete(struct{}{})
	if err == nil {
		t.Error("Delete should fail for non-Modeler")
	}
}

func TestFindAll_Where_Errors(t *testing.T) {
	err := FindAll([]ModelTest{})
	if err == nil {
		t.Error("FindAll should fail for non-pointer slice")
	}
	err = Where([]ModelTest{}, "id = ?", 1)
	if err == nil {
		t.Error("Where should fail for non-pointer slice")
	}
	err = FindAll(&[]int{})
	if err == nil {
		t.Error("FindAll should fail for non-Modeler element")
	}
}

func TestScanRow_Errors(t *testing.T) {
	m := &ModelTest{}
	err := scanRow(123, m)
	if err == nil {
		t.Error("scanRow should fail for unsupported row type")
	}
}
