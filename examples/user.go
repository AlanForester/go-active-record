package main

import (
	"time"

	"go-active-record/activerecord"
)

const (
	minNameLength = 2
	maxNameLength = 50
	minAge        = 18
	maxAge        = 100
	adultAge      = 18
	youngAge      = 30
	middleAge     = 50
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

func (u *User) BeforeCreate() {
	u.SetCreatedAt(time.Now())
	u.SetUpdatedAt(time.Now())
	u.setupValidations()
}

func (u *User) BeforeUpdate() {
	u.SetUpdatedAt(time.Now())
	u.setupValidations()
}

func (u *User) setupValidations() {
	u.PresenceOf("Name")
	u.Length("Name", minNameLength, maxNameLength)
	u.AddValidation("Email", "email", "has invalid format")
	u.Uniqueness("Email")
	u.Numericality("Age", minAge, maxAge)
}

func (u *User) FullName() string {
	return u.Name
}

func (u *User) IsAdult() bool {
	return u.Age >= adultAge
}

func (u *User) AgeGroup() string {
	switch {
	case u.Age < adultAge:
		return "minor"
	case u.Age < youngAge:
		return "young"
	case u.Age < middleAge:
		return "adult"
	default:
		return "senior"
	}
}
