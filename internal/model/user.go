package model

import "time"

// User represents a user in the system.
//
// Fields:
//   - ID: Unique identifier for the user
//   - Name: Full name of the user
//   - Email: Email address of the user
//   - CreatedAt: Timestamp when user was created
//   - UpdatedAt: Timestamp when user was last updated
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
