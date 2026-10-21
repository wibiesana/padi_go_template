package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	models "padi-template/app/Models"
	"github.com/wibiesana/padi_go_core/config"
	"github.com/wibiesana/padi_go_core/response"
	"github.com/wibiesana/padi_go_core/validator"
)

type PasswordResetController struct{}

func NewPasswordResetController() *PasswordResetController {
	return &PasswordResetController{}
}

type ForgotPasswordRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Login    string `json:"login"`
}

type ResetPasswordRequest struct {
	Email                string `json:"email" validate:"required,email"`
	Token                string `json:"token" validate:"required"`
	Password             string `json:"password" validate:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required"`
}

// ForgotPassword generates a reset token for the email
func (c *PasswordResetController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Login)
	}

	if identifier == "" {
		response.UnprocessableEntity(w, map[string]string{
			"email": "Email or username is required",
		}, "Validation failed")
		return
	}

	user, err := (models.User{}).FindByLogin(identifier)
	if err != nil || user == nil || user.ID == 0 {
		// Secretive: don't reveal if user exists for security
		response.Success(w, map[string]string{
			"message": "If your email is registered, you will receive a password reset link shortly.",
		}, "Reset instructions sent")
		return
	}

	// Generate random 32-byte hex token
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Unix() + 3600 // 1 hour expiration

	_ = (models.PasswordReset{}).DeleteByEmail(user.Email)

	resetEntry := models.PasswordReset{}
	resetEntry.Email = user.Email
	resetEntry.Token = token
	resetEntry.ExpiresAt = expiresAt

	if err := resetEntry.Save(); err != nil {
		response.InternalServerError(w, "Failed to generate password reset token")
		return
	}

	resData := map[string]interface{}{
		"message": "If your email is registered, you will receive a password reset link shortly.",
	}

	// In debug mode, expose token for easy API testing
	if config.AppConfig != nil && config.AppConfig.AppDebug {
		resData["debug_token"] = token
	}

	response.Success(w, resData, "Password reset request initiated")
}

// ResetPassword verifies token and updates password
func (c *PasswordResetController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	if req.Password != req.PasswordConfirmation {
		response.UnprocessableEntity(w, map[string]string{"password_confirmation": "Password confirmation does not match"}, "Validation failed")
		return
	}

	resetEntry, err := (models.PasswordReset{}).FindValidToken(req.Email, req.Token)
	if err != nil || resetEntry == nil || resetEntry.ID == 0 {
		response.BadRequest(w, "Invalid or expired password reset token")
		return
	}

	user, err := (models.User{}).FindByEmail(req.Email)
	if err != nil || user == nil || user.ID == 0 {
		response.NotFound(w, "User account not found")
		return
	}

	if err := user.SetPassword(req.Password); err != nil {
		response.InternalServerError(w, "Failed to hash new password")
		return
	}

	if err := user.Save(); err != nil {
		response.InternalServerError(w, "Failed to update password")
		return
	}

	// Clean up tokens for this email
	_ = (models.PasswordReset{}).DeleteByEmail(req.Email)

	response.Success(w, nil, "Password has been safely reset. You can now login with your new password.")
}
