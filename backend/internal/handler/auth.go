package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/auth"
	"github.com/Nurdiansyah0/arff-os-backend/internal/middleware"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
	"github.com/Nurdiansyah0/arff-os-backend/internal/repository"
)

type AuthHandler struct {
	store *repository.Store
}

func NewAuthHandler(store *repository.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

// RegisterAirport handles B2B SaaS onboarding (creates airport + initial airport admin)
func (h *AuthHandler) RegisterAirport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterAirportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.AirportName) == "" {
		http.Error(w, `{"error":"Airport name is required"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, `{"error":"Admin email and password are required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if email already registered
	existing, _ := h.store.GetUserByIdentifier(r.Context(), req.Email)
	if existing != nil {
		http.Error(w, `{"error":"Email is already registered"}`, http.StatusConflict)
		return
	}

	// 1. Create Airport Tenant
	cat := req.Category
	if cat < 1 || cat > 10 {
		cat = 9 // Default category 9 if not specified
	}

	airport := &model.Airport{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.AirportName),
		IATACode:  strings.ToUpper(strings.TrimSpace(req.IATACode)),
		ICAOCode:  strings.ToUpper(strings.TrimSpace(req.ICAOCode)),
		Category:  cat,
		Country:   req.Country,
		Timezone:  "Asia/Jakarta",
		Plan:      "trial",
		Status:    "active",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if airport.Country == "" {
		airport.Country = "ID"
	}

	if err := h.store.CreateAirport(r.Context(), airport); err != nil {
		http.Error(w, `{"error":"Failed to create airport tenant"}`, http.StatusInternalServerError)
		return
	}

	// 2. Hash Password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"Failed to process password encryption"}`, http.StatusInternalServerError)
		return
	}

	// 3. Create Initial Admin User
	adminName := strings.TrimSpace(req.AdminName)
	if adminName == "" {
		adminName = "Chief of ARFF"
	}

	adminUser := &model.User{
		ID:           uuid.New(),
		AirportID:    airport.ID,
		Name:         adminName,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		NIP:          strings.TrimSpace(req.NIP),
		PasswordHash: hashedPassword,
		Role:         "airport_admin",
		Phone:        strings.TrimSpace(req.Phone),
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := h.store.CreateUser(r.Context(), adminUser); err != nil {
		http.Error(w, `{"error":"Failed to create admin user"}`, http.StatusInternalServerError)
		return
	}

	// 4. Generate JWT Token
	token, expiresAt, err := auth.GenerateToken(adminUser.ID, airport.ID, adminUser.Email, adminUser.Role)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate auth token"}`, http.StatusInternalServerError)
		return
	}

	resp := model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      adminUser,
		Airport:   airport,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// Login handles user authentication by Email or NIP + Password
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, `{"error":"Identifier (email/NIP) and password are required"}`, http.StatusBadRequest)
		return
	}

	// Find user
	user, err := h.store.GetUserByIdentifier(r.Context(), identifier)
	if err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		http.Error(w, `{"error":"Account is deactivated"}`, http.StatusForbidden)
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Get airport tenant
	airport, err := h.store.GetAirportByID(r.Context(), user.AirportID)
	if err != nil {
		http.Error(w, `{"error":"Airport tenant not found for user"}`, http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, expiresAt, err := auth.GenerateToken(user.ID, airport.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate auth token"}`, http.StatusInternalServerError)
		return
	}

	resp := model.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      user,
		Airport:   airport,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetMe returns current authenticated user and their airport tenant
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	airport, err := h.store.GetAirportByID(r.Context(), claims.AirportID)
	if err != nil {
		http.Error(w, `{"error":"Airport not found"}`, http.StatusNotFound)
		return
	}

	resp := model.UserProfileResponse{
		User:    user,
		Airport: airport,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
