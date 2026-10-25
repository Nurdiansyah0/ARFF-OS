package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	AirportID uuid.UUID `json:"airport_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "arff-os-super-secret-jwt-key-change-in-production"
	}
	return []byte(secret)
}

// GenerateToken creates a signed JWT token valid for 7 days
func GenerateToken(userID, airportID uuid.UUID, email, role string) (string, time.Time, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:    userID,
		AirportID: airportID,
		Email:     email,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "arff-os-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", expirationTime, err
	}

	return tokenString, expirationTime, nil
}

// ValidateToken parses and validates a token string
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
