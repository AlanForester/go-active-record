package activerecord

import (
	"testing"
)

type ValidationTestModel struct {
	ValidationModel
	Name  string
	Email string
	Age   int
	Role  string
}

func (m *ValidationTestModel) TableName() string { return "validation_test_models" }

func TestValidationModel_PresenceOf(t *testing.T) {
	m := &ValidationTestModel{}
	m.PresenceOf("Name")
	if m.IsValid(m) {
		t.Error("Should be invalid if Name is empty")
	}
	m.Name = "Test"
	if !m.IsValid(m) {
		t.Error("Should be valid if Name is present")
	}
}

func TestValidationModel_Length(t *testing.T) {
	m := &ValidationTestModel{Name: "A"}
	m.Length("Name", 2, 5)
	if m.IsValid(m) {
		t.Error("Should be invalid if Name too short")
	}
	m.Name = "Test"
	if !m.IsValid(m) {
		t.Error("Should be valid if Name length is ok")
	}
	m.Name = "TooLongName"
	if m.IsValid(m) {
		t.Error("Should be invalid if Name too long")
	}
}

func TestValidationModel_Email(t *testing.T) {
	m := &ValidationTestModel{Email: "not-an-email"}
	m.ValidationModel.Email("Email")
	if m.IsValid(m) {
		t.Error("Should be invalid for bad email")
	}
	m.Email = "good@email.com"
	if !m.IsValid(m) {
		t.Error("Should be valid for good email")
	}
}

func TestValidationModel_Numericality(t *testing.T) {
	m := &ValidationTestModel{Age: 10}
	m.Numericality("Age", 18, 100)
	if m.IsValid(m) {
		t.Error("Should be invalid for too small Age")
	}
	m.Age = 25
	if !m.IsValid(m) {
		t.Error("Should be valid for Age in range")
	}
	m.Age = 120
	if m.IsValid(m) {
		t.Error("Should be invalid for too large Age")
	}
}

func TestValidationModel_Format(t *testing.T) {
	m := &ValidationTestModel{Role: "admin"}
	m.Format("Role", "^(admin|user)$")
	if !m.IsValid(m) {
		t.Error("Should be valid for allowed role")
	}
	m.Role = "guest"
	if m.IsValid(m) {
		t.Error("Should be invalid for not allowed role")
	}
}

func TestValidationModel_Errors(t *testing.T) {
	m := &ValidationTestModel{}
	m.PresenceOf("Name")
	m.IsValid(m)
	if len(m.Errors()) == 0 {
		t.Error("Should have errors")
	}
}

func TestValidationModel_HelperMethods(t *testing.T) {
	m := &ValidationTestModel{}
	if !m.isEmpty("") || !m.isEmpty(nil) {
		t.Error("isEmpty should be true for empty string and nil")
	}
	if m.isEmpty("not empty") {
		t.Error("isEmpty should be false for non-empty string")
	}
	if !m.isValidEmail("a@b.com") {
		t.Error("isValidEmail should be true for valid email")
	}
	if m.isValidEmail("bad") {
		t.Error("isValidEmail should be false for invalid email")
	}
	if m.isDuplicate("field", "value") {
		t.Error("isDuplicate should be false (stub)")
	}
	if v, ok := m.toFloat(42); !ok || v != 42 {
		t.Error("toFloat should convert int")
	}
	if v, ok := m.toFloat(3.14); !ok || v != 3.14 {
		t.Error("toFloat should convert float64")
	}
	if !m.matchesPattern("abc", "^abc$") {
		t.Error("matchesPattern should match")
	}
	if m.matchesPattern("def", "^abc$") {
		t.Error("matchesPattern should not match")
	}
}
