package controllers

import (
	"log"
	"net/http"
	"strings"
	"time"

	models "padi-template/app/Models"
	"github.com/wibiesana/padi_go_core/activerecord"
	"github.com/wibiesana/padi_go_core/auth"
	"github.com/wibiesana/padi_go_core/middleware"
	"github.com/wibiesana/padi_go_core/response"
	"github.com/wibiesana/padi_go_core/validator"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"`
}

type FlexibleBool bool

func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		*b = true
	case "false", "0", "no", "null", "":
		*b = false
	default:
		*b = false
	}
	return nil
}

func (b FlexibleBool) Bool() bool {
	return bool(b)
}

type LoginRequest struct {
	Login      string       `json:"login"`
	User       string       `json:"user"`
	Email      string       `json:"email"`
	Username   string       `json:"username"`
	Password   string       `json:"password" validate:"required"`
	RememberMe FlexibleBool `json:"remember_me"`
}

type AuthResponse struct {
	User          models.User `json:"user"`
	Token         string      `json:"token"`
	RememberToken *string     `json:"remember_token,omitempty"`
	ExpiresIn     *int64      `json:"expires_in,omitempty"`
}

// Register creates a new user account and returns JWT token
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	// Ensure either name or username is provided
	if req.Name == "" && req.Username == "" {
		response.UnprocessableEntity(w, map[string]string{
			"username": "Username or name is required",
		}, "Validation failed")
		return
	}

	// Check if email already exists using ActiveRecord FindByEmail
	if existing, err := (models.User{}).FindByEmail(req.Email); err == nil && existing != nil && existing.ID > 0 {
		response.BadRequest(w, "Email is already registered")
		return
	}

	// Check if username already exists if provided
	if req.Username != "" {
		if existing, err := (models.User{}).FindByUsername(req.Username); err == nil && existing != nil && existing.ID > 0 {
			response.BadRequest(w, "Username is already registered")
			return
		}
	}

	user := models.User{}
	if req.Name != "" {
		name := req.Name
		user.Name = &name
	} else if req.Username != "" {
		name := req.Username
		user.Name = &name
	}

	if req.Username != "" {
		username := req.Username
		user.Username = &username
	}
	user.Email = req.Email
	user.Role = req.Role
	if user.Role == "" {
		user.Role = "user"
	}
	user.Status = "active"

	if err := user.SetPassword(req.Password); err != nil {
		response.InternalServerError(w, "Failed to hash password")
		return
	}

	// Save via ActiveRecord Save()
	if err := user.Save(); err != nil {
		response.InternalServerError(w, "Failed to create user account: "+err.Error())
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		response.InternalServerError(w, "User created but token generation failed")
		return
	}

	activerecord.HideFields(&user)

	response.Created(w, AuthResponse{
		User:  user,
		Token: token,
	}, "User registered successfully")
}

func isUserActive(user *models.User) bool {
	if strings.EqualFold(user.Role, "superadmin") {
		return true
	}
	s := strings.TrimSpace(strings.ToLower(user.Status))
	return s == "" || s == "active" || s == "1" || s == "true"
}

// Login authenticates user with email/username & password and returns JWT token
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	loginIdentifier := strings.TrimSpace(req.Login)
	if loginIdentifier == "" {
		loginIdentifier = strings.TrimSpace(req.Username)
	}
	if loginIdentifier == "" {
		loginIdentifier = strings.TrimSpace(req.Email)
	}
	if loginIdentifier == "" {
		loginIdentifier = strings.TrimSpace(req.User)
	}

	if loginIdentifier == "" {
		response.UnprocessableEntity(w, map[string]string{
			"login": "Email or username is required",
		}, "Validation failed")
		return
	}

	user, err := (models.User{}).FindByLogin(loginIdentifier)
	if err != nil || user == nil || user.ID == 0 {
		response.Unauthorized(w, "Invalid email/username or password")
		return
	}

	if !user.VerifyPassword(req.Password) {
		response.Unauthorized(w, "Invalid email/username or password")
		return
	}

	if !isUserActive(user) {
		response.Unauthorized(w, "Account is not active")
		return
	}

	// Update last login
	now := time.Now().Unix()
	user.LastLoginAt = &now

	var rememberToken *string
	var expiresIn *int64
	var token string

	if req.RememberMe.Bool() {
		ttl := int64(365 * 24 * 3600) // 1 year
		expiresIn = &ttl
		rt := auth.GenerateSecureRandomString(32)
		rememberToken = &rt
		user.RememberToken = rememberToken
		token, err = auth.GenerateTokenWithExpiry(user.ID, user.Email, user.Role, time.Duration(ttl)*time.Second)
	} else {
		token, err = auth.GenerateToken(user.ID, user.Email, user.Role)
	}

	if err != nil {
		response.InternalServerError(w, "Failed to generate authentication token")
		return
	}

	if err := user.Save(); err != nil {
		// Log but do not abort — login already succeeded; best-effort persistence
		log.Printf("[WARN] Login: failed to update user %d last_login/remember_token: %v", user.ID, err)
	}

	activerecord.HideFields(user)

	response.Success(w, AuthResponse{
		User:          *user,
		Token:         token,
		RememberToken: rememberToken,
		ExpiresIn:     expiresIn,
	}, "Login successful")
}

// Refresh generates a new access token using remember_token
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RememberToken string `json:"remember_token" validate:"required"`
	}
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	user, err := activerecord.FindBy[models.User]("remember_token", req.RememberToken)
	if err != nil || user == nil || user.ID == 0 {
		response.Unauthorized(w, "Invalid or expired remember token")
		return
	}

	if !isUserActive(user) {
		response.Unauthorized(w, "Your account is inactive. Please contact support.")
		return
	}

	ttl := int64(365 * 24 * 3600)
	token, err := auth.GenerateTokenWithExpiry(user.ID, user.Email, user.Role, time.Duration(ttl)*time.Second)
	if err != nil {
		response.InternalServerError(w, "Failed to generate token")
		return
	}

	now := time.Now().Unix()
	user.LastLoginAt = &now
	if err := user.Save(); err != nil {
		log.Printf("[WARN] Refresh: failed to update user %d last_login: %v", user.ID, err)
	}

	rememberToken := user.RememberToken
	activerecord.HideFields(user)

	response.Success(w, AuthResponse{
		User:          *user,
		Token:         token,
		RememberToken: rememberToken,
		ExpiresIn:     &ttl,
	}, "Token refreshed successfully")
}

// Logout invalidates user session
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	response.Success(w, nil, "Logout successful")
}

// Me returns currently authenticated user profile
func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.JWTClaims)
	if !ok || claims == nil {
		response.Unauthorized(w, "Unauthenticated")
		return
	}

	user, err := (models.User{}).Find(claims.UserID)
	if err != nil {
		response.InternalServerError(w, "Failed to retrieve user profile")
		return
	}
	if user == nil || user.ID == 0 {
		response.NotFound(w, "User profile not found")
		return
	}

	activerecord.HideFields(user)

	response.Success(w, user, "User profile retrieved")
}
