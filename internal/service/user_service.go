package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"api-test/internal/cache"
	"api-test/internal/model"
)

// UserService handles business logic for users.
type UserService struct {
	db    *sql.DB
	cache *cache.Client
}

// New creates a new user service.
func New(db *sql.DB, cache *cache.Client) *UserService {
	return &UserService{
		db:    db,
		cache: cache,
	}
}

// GetAll retrieves all users with caching.
func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	// Check cache first
	cachedUsers, err := s.cache.Get(ctx, "all_users")
	if err == nil {
		var users []model.User
		if err := json.Unmarshal([]byte(cachedUsers), &users); err == nil {
			return users, nil
		}
	}

	// Cache miss - fetch from database
	rows, err := s.db.Query("SELECT id, name, email, created_at, updated_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Cache the result for 5 minutes
	if len(users) > 0 {
		usersJSON, _ := json.Marshal(users)
		_ = s.cache.Set(ctx, "all_users", string(usersJSON), 5*time.Minute)
	}

	return users, nil
}

// GetByID retrieves a single user by ID with caching.
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}

	cacheKey := fmt.Sprintf("user:%s", id)

	// Check cache first
	cachedUser, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var user model.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// Cache miss - fetch from database
	var user model.User
	err = s.db.QueryRow(
		"SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, err
	}

	// Cache the result for 10 minutes
	userJSON, _ := json.Marshal(user)
	_ = s.cache.Set(ctx, cacheKey, string(userJSON), 10*time.Minute)

	return &user, nil
}

// Create creates a new user.
func (s *UserService) Create(ctx context.Context, user *model.User) error {
	if user.Name == "" || user.Email == "" {
		return fmt.Errorf("name and email are required")
	}

	now := time.Now()
	err := s.db.QueryRow(
		"INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Name, user.Email, now, now,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now

	// Invalidate cache
	_ = s.cache.Delete(ctx, "all_users")

	return nil
}
