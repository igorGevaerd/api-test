package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-test/internal/cache"
	"api-test/internal/handler"
	"api-test/internal/model"
	"api-test/internal/service"

	"github.com/go-redis/redis/v8"
)

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	// Using SQLite for testing instead of PostgreSQL
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create users table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := testDB.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return testDB
}

// setupTestRedis initializes a mock Redis client for testing.
func setupTestRedis(t *testing.T) *cache.Client {
	// Connect to Redis (ensure Redis is running for tests)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Logf("Warning: Redis not available for testing: %v", err)
		// Tests will continue but Redis caching tests may be skipped
	}

	return &cache.Client{Underlying: redisClient}
}

// TestHealthCheck tests the health check endpoint.
func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler.Health)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("handler returned unexpected status: got %v want ok", result["status"])
	}
}

// TestGetUsersEmpty tests getting users from an empty database.
func TestGetUsersEmpty(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Mock Redis cache
	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	userHandler.GetAll(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var users []model.User
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if users == nil {
		users = []model.User{}
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

// TestCreateUser tests creating a new user.
func TestCreateUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	user := model.User{
		Name:  "Test User",
		Email: "test@example.com",
	}

	body, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	userHandler.Create(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var createdUser model.User
	if err := json.NewDecoder(rr.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, createdUser.Name)
	}

	if createdUser.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, createdUser.Email)
	}

	if createdUser.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

// TestCreateUserMissingFields tests creating a user with missing required fields.
func TestCreateUserMissingFields(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	testCases := []struct {
		name  string
		user  model.User
		field string
	}{
		{
			name:  "missing name",
			user:  model.User{Email: "test@example.com"},
			field: "name",
		},
		{
			name:  "missing email",
			user:  model.User{Name: "Test User"},
			field: "email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.user)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.Create(rr, req)

			if status := rr.Code; status != http.StatusBadRequest {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
			}

			var errResponse map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&errResponse); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if _, exists := errResponse["error"]; !exists {
				t.Error("expected error field in response")
			}
		})
	}
}

// TestCreateUserInvalidJSON tests creating a user with invalid JSON.
func TestCreateUserInvalidJSON(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	userHandler.Create(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// TestGetUserNotFound tests getting a non-existent user.
func TestGetUserNotFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	req, err := http.NewRequest("GET", "/user?id=999", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	userHandler.GetByID(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var errResponse map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if errResponse["error"] != "user not found" {
		t.Errorf("expected 'user not found', got %s", errResponse["error"])
	}
}

// TestGetUserMissingID tests getting a user without ID parameter.
func TestGetUserMissingID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	req, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	userHandler.GetByID(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var errResponse map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if errResponse["error"] != "id parameter is required" {
		t.Errorf("expected 'id parameter is required', got %s", errResponse["error"])
	}
}

// TestGetUserFound tests getting an existing user.
func TestGetUserFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	// Insert test user
	insertSQL := `INSERT INTO users (name, email, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, err := testDB.Exec(insertSQL, "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	userID, _ := result.LastInsertId()

	req, err := http.NewRequest("GET", "/user?id="+string(rune(userID)), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	userHandler.GetByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var user model.User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", user.Email)
	}
}

// TestContentType tests that responses have correct Content-Type header.
func TestContentType(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	mockCache := &cache.Client{Underlying: redis.NewClient(&redis.Options{Addr: "localhost:6379"})}

	userService := service.New(testDB, mockCache)
	userHandler := handler.New(userService)

	testCases := []struct {
		name    string
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{
			name:   "GET /users",
			method: "GET",
			path:   "/users",
			handler: func(w http.ResponseWriter, r *http.Request) {
				userHandler.GetAll(w, r)
			},
		},
		{
			name:   "GET /health",
			method: "GET",
			path:   "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				handler.Health(w, r)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			tc.handler.ServeHTTP(rr, req)

			expected := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expected {
				t.Errorf("expected Content-Type %s, got %s", expected, ct)
			}
		})
	}
}
