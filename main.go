package main

import (
	"fmt"
	"log"

	"go-active-record/activerecord"
)

const (
	defaultAge = 30
	youngAge   = 25
)

// User model - example of using Active Record.
type User struct {
	activerecord.ActiveRecordModel
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Age   int    `db:"age" json:"age"`
}

// TableName returns the name of the table for the model.
func (u *User) TableName() string {
	return "users"
}

func main() {
	// Initialize connection to the database
	db, err := activerecord.Connect("postgres",
		"host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set connection as global
	activerecord.SetConnection(db, "postgres")

	// Examples of using Active Record

	// Creating a new user
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   defaultAge,
	}

	if err := user.Create(); err != nil {
		log.Printf("Failed to create user: %v", err)
	}

	// Finding a user by ID
	foundUser := &User{}
	if err := activerecord.Find(foundUser, 1); err != nil {
		log.Printf("Failed to find user: %v", err)
	} else {
		fmt.Printf("Found user: %+v\n", foundUser)
	}

	// Finding all users
	var users []User
	if err := activerecord.FindAll(&users); err != nil {
		log.Printf("Failed to find users: %v", err)
	} else {
		fmt.Printf("Found users: %d\n", len(users))
	}

	// Finding with conditions
	var youngUsers []User
	if err := activerecord.Where(&youngUsers, "age < ?", youngAge); err != nil {
		log.Printf("Failed to find young users: %v", err)
	} else {
		fmt.Printf("Young users: %d\n", len(youngUsers))
	}

	// Updating a user
	if foundUser.IsPersisted() {
		foundUser.Age = 31
		if err := foundUser.Update(); err != nil {
			log.Printf("Failed to update user: %v", err)
		}
	}

	// Deleting a user
	if foundUser.IsPersisted() {
		if err := foundUser.Delete(); err != nil {
			log.Printf("Failed to delete user: %v", err)
		}
	}
}
