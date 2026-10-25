package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	AirportID    uuid.UUID `json:"airport_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	NIP          string    `json:"nip,omitempty"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Phone        string    `json:"phone,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	// Supports Login by Email or NIP
	Identifier string `json:"identifier"` // Email or NIP
	Password   string `json:"password"`
}

type AuthResponse struct {
	Token     string   `json:"token"`
	ExpiresAt string   `json:"expires_at"`
	User      *User    `json:"user"`
	Airport   *Airport `json:"airport"`
}

type UserProfileResponse struct {
	User    *User    `json:"user"`
	Airport *Airport `json:"airport"`
}
