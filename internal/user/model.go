package user

import "time"

type User struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	Name            string    `json:"name"`
	IsEmailVerified bool      `json:"is_email_verified"`
	GoogleID        string    `json:"google_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
