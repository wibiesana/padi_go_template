package models

import (
	"strings"
	"time"

	"github.com/wibiesana/padi_go_core/activerecord"
	"github.com/wibiesana/padi_go_core/auth"
)

type User struct {
	ID              uint       `db:"id" json:"id"`
	Name            *string    `db:"name" json:"name,omitempty"`
	Username        *string    `db:"username" json:"username,omitempty"`
	Email           string     `db:"email" json:"email" validate:"required,email"`
	Password        string     `db:"password" json:"-" validate:"required"`
	Role            string     `db:"role" json:"role"`
	Status          string     `db:"status" json:"status,omitempty"`
	EmailVerifiedAt *int64     `db:"email_verified_at" json:"email_verified_at,omitempty"`
	RememberToken   *string    `db:"remember_token" json:"-"`
	LastLoginAt     *int64     `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedBy       *uint      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy       *uint      `db:"updated_by" json:"updated_by,omitempty"`
	CreatedAt       *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// Hidden returns sensitive fields to be hidden
func (User) Hidden() []string {
	return []string{"password", "remember_token"}
}

// Find finds user by ID
func (User) Find(id interface{}) (*User, error) {
	return activerecord.Find[User](id)
}

// FindOrFail finds user by ID or returns error if not found
func (User) FindOrFail(id interface{}) (*User, error) {
	return activerecord.FindOrFail[User](id)
}

// All retrieves all users
func (User) All(columns ...string) ([]User, error) {
	return activerecord.All[User](columns...)
}

// FindByEmail finds user by email
func (User) FindByEmail(email string) (*User, error) {
	return activerecord.FindBy[User]("email", email)
}

// FindByUsername finds user by username
func (User) FindByUsername(username string) (*User, error) {
	return activerecord.FindBy[User]("username", username)
}

// FindByLogin finds user by email, username, or name
func (User) FindByLogin(login string) (*User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, nil
	}

	// 1. Check email
	item, err := activerecord.NewModelQuery[User]().Where("email", login).First()
	if err == nil && item != nil && item.ID > 0 {
		return item, nil
	}

	// Check lowercase email if different
	lower := strings.ToLower(login)
	if lower != login {
		item, err = activerecord.NewModelQuery[User]().Where("email", lower).First()
		if err == nil && item != nil && item.ID > 0 {
			return item, nil
		}
	}

	// 2. Check username
	item, err = activerecord.NewModelQuery[User]().Where("username", login).First()
	if err == nil && item != nil && item.ID > 0 {
		return item, nil
	}

	// Check lowercase username if different
	if lower != login {
		item, err = activerecord.NewModelQuery[User]().Where("username", lower).First()
		if err == nil && item != nil && item.ID > 0 {
			return item, nil
		}
	}

	// 3. Check name
	item, err = activerecord.NewModelQuery[User]().Where("name", login).First()
	if err == nil && item != nil && item.ID > 0 {
		return item, nil
	}

	return nil, nil
}

// Save saves or updates user
func (u *User) Save() error {
	return activerecord.Save(u)
}

// Delete removes user
func (u *User) Delete() error {
	return activerecord.DeleteModel(u)
}

// SetPassword hashes and sets user password
func (u *User) SetPassword(plainPassword string) error {
	hash, err := auth.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

// VerifyPassword checks if plain password matches hash
func (u *User) VerifyPassword(plainPassword string) bool {
	return auth.CheckPasswordHash(plainPassword, u.Password)
}
