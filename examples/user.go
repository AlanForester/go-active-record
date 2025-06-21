package main

import (
	"time"

	"go-active-record/activerecord"
)

type User struct {
	activerecord.ValidationModel
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Age       int       `db:"age" json:"age"`
	Mentor    *User     `db:"mentor_id" json:"mentor_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate() error {
	u.setupValidations()
	return nil
}

func (u *User) BeforeUpdate() error {
	u.setupValidations()
	return nil
}

func (u *User) setupValidations() {
	u.PresenceOf("Name")
	u.Length("Name", 2, 50)
	u.AddValidation("Email", "email", "has invalid format")
	u.Uniqueness("Email")
	u.Numericality("Age", 18, 100)
}

func (u *User) FullName() string {
	return u.Name
}

func (u *User) IsAdult() bool {
	return u.Age >= 18
}

func (u *User) AgeGroup() string {
	switch {
	case u.Age < 18:
		return "minor"
	case u.Age < 30:
		return "young"
	case u.Age < 50:
		return "adult"
	default:
		return "senior"
	}
}
