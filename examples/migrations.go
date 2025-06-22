package main

import (
	"github.com/Forester-Co/go-active-record/activerecord"
)

const (
	createUsersTableVersion = 20231201000001
	createPostsTableVersion = 20231201000002
	createTagsTableVersion  = 20231201000003
	addUserRoleVersion      = 20231201000003
)

// CreateUsersTable migration for creating users table.
type CreateUsersTable struct {
	activerecord.Migration
}

func (m *CreateUsersTable) Version() int64 {
	return createUsersTableVersion
}

func (m *CreateUsersTable) Up() error {
	return activerecord.CreateTable("users", func(t *activerecord.TableBuilder) {
		t.Column("id", "SERIAL", "PRIMARY KEY")
		t.Column("name", "VARCHAR(255)", "NOT NULL")
		t.Column("email", "VARCHAR(255)", "UNIQUE", "NOT NULL")
		t.Column("age", "INTEGER")
		t.Timestamps()
		t.Index("email")
	})
}

func (m *CreateUsersTable) Down() error {
	return activerecord.DropTable("users")
}

// CreatePostsTable migration for creating posts table.
type CreatePostsTable struct {
	activerecord.Migration
}

func (m *CreatePostsTable) Version() int64 {
	return createPostsTableVersion
}

func (m *CreatePostsTable) Up() error {
	return activerecord.CreateTable("posts", func(t *activerecord.TableBuilder) {
		t.Column("id", "SERIAL", "PRIMARY KEY")
		t.Column("title", "VARCHAR(255)", "NOT NULL")
		t.Column("content", "TEXT")
		t.Column("user_id", "INTEGER", "NOT NULL")
		t.Column("published", "BOOLEAN", "DEFAULT FALSE")
		t.Timestamps()
		t.Index("user_id")
		t.Index("published")
	})
}

func (m *CreatePostsTable) Down() error {
	return activerecord.DropTable("posts")
}

// AddUserRole migration for adding user role.
type AddUserRole struct {
	activerecord.Migration
}

func (m *AddUserRole) Version() int64 {
	return addUserRoleVersion
}

func (m *AddUserRole) Up() error {
	query := "ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user'"
	_, err := activerecord.GetConnection().Exec(query)
	return err
}

func (m *AddUserRole) Down() error {
	query := "ALTER TABLE users DROP COLUMN role"
	_, err := activerecord.GetConnection().Exec(query)
	return err
}

// CreateTagsTable migration for creating tags table.
type CreateTagsTable struct {
	activerecord.Migration
}

func (m *CreateTagsTable) Version() int64 {
	return createTagsTableVersion
}

func (m *CreateTagsTable) Up() error {
	return activerecord.CreateTable("tags", func(t *activerecord.TableBuilder) {
		t.Column("id", "SERIAL", "PRIMARY KEY")
		t.Column("name", "VARCHAR(255)", "NOT NULL")
		t.Timestamps()
		t.Index("name")
	})
}

func (m *CreateTagsTable) Down() error {
	return activerecord.DropTable("tags")
}
