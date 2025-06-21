package activerecord

import (
	"reflect"
	"testing"
)

type DummyModel struct{ BaseModel }

func (d *DummyModel) TableName() string { return "dummy" }

func TestActiveRecordModel_Create_Update_Delete_Save(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE dummy (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	m := &ActiveRecordModel{}
	m.SetID(nil)
	err := m.Create()
	if err == nil {
		t.Error("Create should fail for non-Modeler")
	}
	d := &DummyModel{}
	if err := Create(d); err != nil {
		t.Errorf("Create failed: %v", err)
	}
	d.SetID(1)
	ar := &ActiveRecordModel{}
	ar.SetID(1)
	ar.BaseModel = d.BaseModel
	err = ar.Update()
	if err == nil {
		t.Error("Update should fail for non-Modeler")
	}
	err = ar.Delete()
	if err == nil {
		t.Error("Delete should fail for non-Modeler")
	}
}

func TestActiveRecordModel_Save_IsNewRecord_IsPersisted(t *testing.T) {
	m := &ActiveRecordModel{}
	if !m.IsNewRecord() {
		t.Error("Should be new record")
	}
	m.SetID(1)
	if m.IsNewRecord() {
		t.Error("Should not be new record")
	}
	if !m.IsPersisted() {
		t.Error("Should be persisted")
	}
}

func TestActiveRecordModel_Touch_Reload_Destroy(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE dummy (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	d := &DummyModel{}
	if err := Create(d); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	ar := &ActiveRecordModel{}
	ar.SetID(d.GetID())
	ar.BaseModel = d.BaseModel
	err := ar.Touch()
	if err == nil {
		t.Error("Touch should fail for non-Modeler")
	}
	ar.SetID(nil)
	if ar.Reload() != nil {
		t.Error("Reload should return nil for new record")
	}
	ar.SetID(1)
	if ar.Destroy() {
		t.Error("Destroy should fail for non-Modeler")
	}
}

func TestActiveRecordModel_Find_Where(t *testing.T) {
	db, _ := Connect("sqlite3", ":memory:")
	SetConnection(db, "sqlite3")
	if _, err := db.Exec(`CREATE TABLE dummy (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		created_at TIMESTAMP, 
		updated_at TIMESTAMP
	)`); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	d := &DummyModel{}
	if err := Create(d); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	ar := &ActiveRecordModel{}
	ar.SetID(d.GetID())
	ar.BaseModel = d.BaseModel
	err := ar.Find(1)
	if err == nil {
		t.Error("Find should fail for non-Modeler")
	}
	res, err := ar.Where("id = ?", 1)
	if err == nil {
		t.Error("Where should fail for non-Modeler")
	}
	t.Logf("res type: %T, value: %#v", res, res)
	if res == nil {
		t.Error("Where should return a slice, got nil")
	} else if reflect.ValueOf(res).Kind() != reflect.Slice {
		t.Error("Where should return a slice")
	}
}
