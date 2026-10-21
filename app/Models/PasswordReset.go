package models

import (
	"time"

	"github.com/wibiesana/padi_go_core/activerecord"
	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/query"
)

// PasswordReset is a standalone core model (Not generated in Base, not overwritten)
type PasswordReset struct {
	ID        uint       `db:"id" json:"id"`
	Email     string     `db:"email" json:"email" validate:"required,email"`
	Token     string     `db:"token" json:"token" validate:"required"`
	ExpiresAt int64      `db:"expires_at" json:"expires_at"`
	CreatedBy *uint      `db:"created_by" json:"created_by,omitempty"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

// FindValidToken retrieves an unexpired token record for an email
func (PasswordReset) FindValidToken(email, token string) (*PasswordReset, error) {
	var pr PasswordReset
	now := time.Now().Unix()
	err := query.New("password_resets").
		Where("email", email).
		Where("token", token).
		Where("expires_at", ">", now).
		First(&pr)
	return &pr, err
}

// DeleteByEmail removes existing tokens for an email
func (PasswordReset) DeleteByEmail(email string) error {
	db := database.GetDB()
	driver := database.GetDriver()
	delSQL := "DELETE FROM password_resets WHERE email = ?"
	if driver == "postgres" {
		delSQL = "DELETE FROM password_resets WHERE email = $1"
	}
	_, err := db.Exec(delSQL, email)
	return err
}

// Save inserts password reset record
func (pr *PasswordReset) Save() error {
	return activerecord.Save(pr)
}
