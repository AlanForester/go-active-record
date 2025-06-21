package main

import (
	"go-active-record/activerecord"
)

// Post model with validations
type Post struct {
	activerecord.ValidationModel
	Title   string `db:"title" json:"title"`
	Content string `db:"content" json:"content"`
	UserID  int    `db:"user_id" json:"user_id"`
}

// TableName returns the table name
func (p *Post) TableName() string {
	return "posts"
}

// SetupValidations sets up validations for the model
func (p *Post) SetupValidations() {
	// Required fields
	p.PresenceOf("Title")
	p.PresenceOf("Content")
	p.PresenceOf("UserID")

	// Title length validation
	p.Length("Title", 5, 200)

	// Content length validation
	p.Length("Content", 10, 10000)

	// UserID validation
	p.Numericality("UserID", 1, 999999)
}

// Example usage of validations for Post
// func main() {
// 	// Example post validation
// 	post := &Post{
// 		Title:   "Hi",    // Too short title
// 		Content: "Short", // Too short content
// 		UserID:  0,       // Invalid UserID
// 	}
//
// 	post.SetupValidations()
//
// 	if !post.IsValid() {
// 		fmt.Println("Post validation errors:")
// 		for _, err := range post.Errors() {
// 			fmt.Printf("- %s: %s\n", err.Field, err.Message)
// 		}
// 	}
// }
