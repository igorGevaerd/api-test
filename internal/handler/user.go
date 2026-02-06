package handler

import (
	"encoding/json"
	"net/http"

	"api-test/internal/model"
	"api-test/internal/service"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	service *service.UserService
}

// New creates a new user handler.
func New(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetAll handles GET requests to /users and returns all users.
//
// HTTP Response:
//   - Status 200: Successfully returns a JSON array of users
//   - Status 500: Server error
//   - Content-Type: application/json
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	users, err := h.service.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if len(users) == 0 {
		_ = json.NewEncoder(w).Encode([]model.User{})
	} else {
		_ = json.NewEncoder(w).Encode(users)
	}
}

// GetByID handles GET requests to /user and returns a single user by ID.
//
// Query Parameters:
//   - id: The ID of the user to retrieve (required)
//
// HTTP Response:
//   - Status 200: Successfully returns the requested user as JSON
//   - Status 400: Missing or invalid ID parameter
//   - Status 404: User not found
//   - Status 500: Server error
//   - Content-Type: application/json
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "id parameter is required"})
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

// Create handles POST requests to /users and creates a new user.
//
// Request Body:
//
//	Expects JSON with fields: name, email
//	Example: {"name":"Charlie","email":"charlie@example.com"}
//
// HTTP Response:
//   - Status 201: Successfully created user, returns the created user as JSON
//   - Status 400: Invalid request body or missing required fields
//   - Status 500: Server error
//   - Content-Type: application/json
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.Create(r.Context(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}
