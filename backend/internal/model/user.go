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
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // Email or NIP
	Password   string `json:"password"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type ResendOTPRequest struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	Message          string `json:"message"`
	Email            string `json:"email"`
	ExpiresInMinutes int    `json:"expires_in_minutes"`
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
