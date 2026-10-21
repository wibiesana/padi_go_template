package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	routes "padi-template/app/Routes"
	_ "padi-template/database/migrations"
	"github.com/wibiesana/padi_go_core/config"
	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/migrator"
	"github.com/wibiesana/padi_go_core/router"
)

func setupTestApp(t *testing.T) *router.Router {
	cfg := &config.Config{
		AppName:         "Padi Test API",
		AppEnv:          "testing",
		AppPort:         "8080",
		AppDebug:        true,
		AppKey:          "testing-key",
		JWTSecret:       "test-secret-32-chars-long-minimum-jwt",
		JWTExpiration:   1,
		DBConnection:    "sqlite",
		DBDatabase:      ":memory:",
		CorsOrigins:     []string{"*"},
		RateLimitReqs:   1000,
		RateLimitWindow: 60,
	}
	config.AppConfig = cfg

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	if err := migrator.RunPending(db); err != nil {
		t.Fatalf("Failed to run migrations on test db: %v", err)
	}

	r := router.New(cfg)
	routes.RegisterRoutes(r)
	return r
}

func TestAuthAndUserEndpoints(t *testing.T) {
	r := setupTestApp(t)

	// 1. Test Register with username and email
	regPayload := map[string]string{
		"name":     "John Doe",
		"username": "johndoe",
		"email":    "john@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on register, got %d. Body: %s", w.Code, w.Body.String())
	}

	var regResponse struct {
		Success bool `json:"success"`
		Item    struct {
			Token string `json:"token"`
		} `json:"item"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &regResponse); err != nil || regResponse.Item.Token == "" {
		t.Fatalf("Failed to parse token from register response: %v", err)
	}
	token := regResponse.Item.Token

	// 2a. Test Login with Email field
	loginPayloadEmail := map[string]string{
		"email":    "john@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayloadEmail)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with email, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2b. Test Login with Username field
	loginPayloadUsername := map[string]string{
		"username": "johndoe",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayloadUsername)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with username, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2c. Test Login with Login field (providing email)
	loginPayloadLoginEmail := map[string]string{
		"login":    "john@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayloadLoginEmail)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with login=email, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2d. Test Login with Login field (providing username)
	loginPayloadLoginUser := map[string]string{
		"login":    "johndoe",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayloadLoginUser)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with login=username, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2e. Test Login with email in username field
	loginPayloadEmailInUser := map[string]string{
		"username": "john@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginPayloadEmailInUser)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with email in username field, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2f. Test Login with invalid password
	loginPayloadBadPass := map[string]string{
		"username": "johndoe",
		"password": "wrongpassword",
	}
	body, _ = json.Marshal(loginPayloadBadPass)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized on invalid password, got %d", w.Code)
	}

	// 2g. Test Login with remember_me = true
	loginPayloadRemember := map[string]interface{}{
		"username":    "johndoe",
		"password":    "password123",
		"remember_me": true,
	}
	body, _ = json.Marshal(loginPayloadRemember)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login with remember_me=true, got %d. Body: %s", w.Code, w.Body.String())
	}
	var loginRemResp struct {
		Success bool `json:"success"`
		Item    struct {
			Token         string `json:"token"`
			RememberToken string `json:"remember_token"`
			ExpiresIn     int64  `json:"expires_in"`
		} `json:"item"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginRemResp); err != nil || loginRemResp.Item.RememberToken == "" {
		t.Fatalf("Failed to parse remember_token from response: %v, body: %s", err, w.Body.String())
	}
	if loginRemResp.Item.ExpiresIn != 365*24*3600 {
		t.Fatalf("Expected expires_in 31536000, got %d", loginRemResp.Item.ExpiresIn)
	}

	// 2h. Test Refresh with remember_token
	refreshPayload := map[string]string{
		"remember_token": loginRemResp.Item.RememberToken,
	}
	body, _ = json.Marshal(refreshPayload)
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /auth/refresh with remember_token, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 3. Test Me with Bearer Token
	req = httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /auth/me, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. Test User CRUD Index
	req = httptest.NewRequest(http.MethodGet, "/users?page=1&per_page=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /users, got %d. Body: %s", w.Code, w.Body.String())
	}
}
