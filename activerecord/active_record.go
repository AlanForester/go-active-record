package activerecord

import (
	"errors"
	"reflect"
	"time"
)

// ActiveRecord interface for instance methods of the model
type ActiveRecord interface {
	Modeler
	Create() error
	Update() error
	Delete() error
	Save() error
	IsNewRecord() bool
	IsPersisted() bool
}

// ActiveRecordModel base model with Active Record methods
type ActiveRecordModel struct {
	BaseModel // anonymous embedding
}

// Create creates a new record in the database
func (m *ActiveRecordModel) Create() error {
	return Create(m)
}

// Update updates a record in the database
func (m *ActiveRecordModel) Update() error {
	model, ok := any(m).(Modeler)
	if !ok {
		return ErrNotModeler
	}
	return Update(model)
}

// Delete deletes a record from the database
func (m *ActiveRecordModel) Delete() error {
	model, ok := any(m).(Modeler)
	if !ok {
		return ErrNotModeler
	}
	return Delete(model)
}

// Save saves a record (creates or updates)
func (m *ActiveRecordModel) Save() error {
	if m.IsNewRecord() {
		return m.Create()
	}
	return m.Update()
}

// IsNewRecord checks if a record is new
func (m *ActiveRecordModel) IsNewRecord() bool {
	return m.GetID() == nil || m.GetID() == 0
}

// IsPersisted checks if a record is saved in the database
func (m *ActiveRecordModel) IsPersisted() bool {
	return !m.IsNewRecord()
}

// Touch updates timestamps
func (m *ActiveRecordModel) Touch() error {
	m.SetUpdatedAt(time.Now())
	model, ok := any(m).(Modeler)
	if !ok {
		return ErrNotModeler
	}
	return Update(model)
}

// Reload reloads data from the database
func (m *ActiveRecordModel) Reload() error {
	if m.IsNewRecord() {
		return nil
	}
	model, ok := any(m).(Modeler)
	if !ok {
		return ErrNotModeler
	}
	return Find(model, m.GetID())
}

// Destroy deletes a record and returns true if successful
func (m *ActiveRecordModel) Destroy() bool {
	if m.IsNewRecord() {
		return false
	}
	if err := m.Delete(); err != nil {
		return false
	}
	return true
}

// Modeler interface methods
func (m *ActiveRecordModel) GetID() interface{}       { return m.BaseModel.GetID() }
func (m *ActiveRecordModel) SetID(id interface{})     { m.BaseModel.SetID(id) }
func (m *ActiveRecordModel) GetCreatedAt() time.Time  { return m.BaseModel.GetCreatedAt() }
func (m *ActiveRecordModel) SetCreatedAt(t time.Time) { m.BaseModel.SetCreatedAt(t) }
func (m *ActiveRecordModel) GetUpdatedAt() time.Time  { return m.BaseModel.GetUpdatedAt() }
func (m *ActiveRecordModel) SetUpdatedAt(t time.Time) { m.BaseModel.SetUpdatedAt(t) }

// Find fills the receiver by id
func (m *ActiveRecordModel) Find(id interface{}) error {
	_, ok := any(m).(Modeler)
	if !ok {
		return ErrNotModeler
	}
	return Find(m, id)
}

// Where returns a slice of the receiver's type matching the query
func (m *ActiveRecordModel) Where(query string, args ...interface{}) (interface{}, error) {
	typeOf := reflect.TypeOf(m).Elem()
	sliceType := reflect.SliceOf(typeOf)
	slicePtr := reflect.New(sliceType)
	err := Where(slicePtr.Interface(), query, args...)
	return slicePtr.Elem().Interface(), err
}

var ErrNotModeler = errors.New("receiver does not implement Modeler")
