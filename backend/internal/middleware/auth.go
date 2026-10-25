package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Nurdiansyah0/arff-os-backend/internal/auth"
)

type contextKey string

const (
	UserContextKey contextKey = "arff_user_claims"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header missing"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error": "Invalid Authorization header format. Expected Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Store claims in request context for downstream handlers
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims extracts the user claims from request context
func GetUserClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}
