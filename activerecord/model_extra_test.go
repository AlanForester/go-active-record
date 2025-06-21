package activerecord

import (
	"reflect"
	"testing"
	"time"
)

type ModelTest struct {
	BaseModel
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (m *ModelTest) TableName() string { return "model_tests" }

func TestBaseModelMethods(t *testing.T) {
	b := &BaseModel{}
	b.SetID(int64(42))
	if b.GetID() != 42 {
		t.Error("SetID/GetID failed")
	}
	now := time.Now()
	b.SetCreatedAt(now)
	b.SetUpdatedAt(now)
	if !b.GetCreatedAt().Equal(now) || !b.GetUpdatedAt().Equal(now) {
		t.Error("Set/Get CreatedAt/UpdatedAt failed")
	}
}

func TestSetTimestampsDeep(t *testing.T) {
	m := &ModelTest{}
	setTimestampsDeep(reflect.ValueOf(m), time.Now())
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		t.Error("setTimestampsDeep should set timestamps")
	}
}

func TestGetFieldsAndValues(t *testing.T) {
	m := &ModelTest{Name: "A", Age: 1}
	fields, values := getFieldsAndValues(m, false)
	if len(fields) == 0 || len(values) == 0 {
		t.Error("getFieldsAndValues should return fields and values")
	}
}

func TestCreate_Find_Update_Delete(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	db.Exec(`CREATE TABLE model_tests (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER, created_at TIMESTAMP, updated_at TIMESTAMP)`)
	m := &ModelTest{Name: "Test", Age: 10}
	err := Create(m)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	m.Age = 20
	err = Update(m)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	found := &ModelTest{}
	err = Find(found, m.GetID())
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	err = Delete(m)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func TestFindAll_Where(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	db.Exec(`CREATE TABLE model_tests (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER, created_at TIMESTAMP, updated_at TIMESTAMP)`)
	Create(&ModelTest{Name: "A", Age: 1})
	Create(&ModelTest{Name: "B", Age: 2})
	var all []ModelTest
	err := FindAll(&all)
	if err != nil || len(all) < 2 {
		t.Error("FindAll failed or not enough records")
	}
	var filtered []ModelTest
	err = Where(&filtered, "age = ?", 2)
	if err != nil || len(filtered) != 1 {
		t.Error("Where failed or wrong number of records")
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
	err := scanRow(123, m, []string{"id"})
	if err == nil {
		t.Error("scanRow should fail for unsupported row type")
	}
}
