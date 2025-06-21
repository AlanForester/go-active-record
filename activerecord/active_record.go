package activerecord

import (
	"errors"
	"fmt"
	"time"
)

// ActiveRecord interface for instance methods of the model.
type ActiveRecord interface {
	Modeler
	Create() error
	Update() error
	Delete() error
	Save() error
	IsNewRecord() bool
	IsPersisted() bool
	Touch() error
	Reload() error
	Destroy() bool
}

// ActiveRecordModel base model with Active Record methods.
type ActiveRecordModel struct {
	BaseModel
}

// Create creates a new record in the database.
func (m *ActiveRecordModel) Create() error {
	return Create(m)
}

// Update updates a record in the database.
func (m *ActiveRecordModel) Update() error {
	return Update(m)
}

// Delete deletes a record from the database.
func (m *ActiveRecordModel) Delete() error {
	return Delete(m)
}

// Save saves a record (creates or updates).
func (m *ActiveRecordModel) Save() error {
	if m.IsNewRecord() {
		return m.Create()
	}
	return m.Update()
}

// IsNewRecord checks if a record is new.
func (m *ActiveRecordModel) IsNewRecord() bool {
	return m.GetID() == nil
}

// IsPersisted checks if a record is saved in the database.
func (m *ActiveRecordModel) IsPersisted() bool {
	return !m.IsNewRecord()
}

// Touch updates timestamps.
func (m *ActiveRecordModel) Touch() error {
	now := time.Now()
	m.SetUpdatedAt(now)
	if m.IsNewRecord() {
		m.SetCreatedAt(now)
	}
	return m.Update()
}

// Reload reloads data from the database.
func (m *ActiveRecordModel) Reload() error {
	if m.IsNewRecord() {
		return nil
	}
	return Find(m, m.GetID())
}

// Destroy deletes a record and returns true if successful.
func (m *ActiveRecordModel) Destroy() bool {
	err := m.Delete()
	return err == nil
}

// Modeler interface methods.
func (m *ActiveRecordModel) GetID() interface{}       { return m.BaseModel.GetID() }
func (m *ActiveRecordModel) SetID(id interface{})     { m.BaseModel.SetID(id) }
func (m *ActiveRecordModel) GetCreatedAt() time.Time  { return m.BaseModel.GetCreatedAt() }
func (m *ActiveRecordModel) SetCreatedAt(t time.Time) { m.BaseModel.SetCreatedAt(t) }
func (m *ActiveRecordModel) GetUpdatedAt() time.Time  { return m.BaseModel.GetUpdatedAt() }
func (m *ActiveRecordModel) SetUpdatedAt(t time.Time) { m.BaseModel.SetUpdatedAt(t) }

// Find fills the receiver by id.
func (m *ActiveRecordModel) Find(id interface{}) error {
	return Find(m, id)
}

// Where returns a slice of the receiver's type matching the query.
func (m *ActiveRecordModel) Where(query string, args ...interface{}) (interface{}, error) {
	// This would need to be implemented based on the specific model type
	return []interface{}{nil}, fmt.Errorf("Where method not implemented for generic ActiveRecordModel")
}

var ErrNotModeler = errors.New("receiver does not implement Modeler")
