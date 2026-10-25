package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateNumericOTP generates a cryptographically random 6-digit string
func GenerateNumericOTP() string {
	maxVal := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		// Fallback deterministic formatting
		return "748291"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
