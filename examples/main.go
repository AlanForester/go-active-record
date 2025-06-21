package main

import (
	"fmt"
	"log"

	"go-active-record/activerecord"
)

const (
	defaultAge    = 30
	youngAgeLimit = 35
)

func main() {
	_, err := activerecord.Connect("sqlite3", "./test.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	migrator := activerecord.NewMigrator()
	migrations := []activerecord.MigrationInterface{&CreateUsersTable{}}

	if err := migrator.Migrate(migrations); err != nil {
		activerecord.Close()
		log.Fatal("Failed to run migration:", err)
	}

	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   defaultAge,
	}

	if !user.IsValid(user) {
		fmt.Println("Validation errors:", user.Errors())
		activerecord.Close()
		return
	}

	err = user.Create()
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("Created user: %+v\n", user)

	var foundUser User
	err = activerecord.Find(&foundUser, user.GetID())
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to find user:", err)
	}

	fmt.Printf("Found user: %+v\n", foundUser)

	foundUser.Age = 31
	err = foundUser.Update()
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to update user:", err)
	}

	fmt.Printf("Updated user: %+v\n", foundUser)

	var allUsers []User
	err = activerecord.FindAll(&allUsers)
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to find all users:", err)
	}

	fmt.Printf("All users: %+v\n", allUsers)

	var youngUsers []User
	err = activerecord.Where(&youngUsers, "age < ?", youngAgeLimit)
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to find young users:", err)
	}

	fmt.Printf("Young users: %+v\n", youngUsers)

	err = foundUser.Delete()
	if err != nil {
		activerecord.Close()
		log.Fatal("Failed to delete user:", err)
	}

	fmt.Println("User deleted successfully")
	activerecord.Close()
}
