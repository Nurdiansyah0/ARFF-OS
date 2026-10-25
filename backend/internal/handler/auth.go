package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/auth"
	"github.com/Nurdiansyah0/arff-os-backend/internal/mailer"
	"github.com/Nurdiansyah0/arff-os-backend/internal/middleware"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
	"github.com/Nurdiansyah0/arff-os-backend/internal/repository"
)

type AuthHandler struct {
	store  *repository.Store
	mailer *mailer.Mailer
}

func NewAuthHandler(store *repository.Store, mailer *mailer.Mailer) *AuthHandler {
	return &AuthHandler{
		store:  store,
		mailer: mailer,
	}
}

// RegisterAirport handles B2B SaaS onboarding (creates airport + initial admin + sends email OTP)
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
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, `{"error":"Admin email and password are required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if email already registered
	existing, _ := h.store.GetUserByIdentifier(r.Context(), email)
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

	// 3. Create Initial Admin User (Unverified initially)
	adminName := strings.TrimSpace(req.AdminName)
	if adminName == "" {
		adminName = "Chief of ARFF"
	}

	adminUser := &model.User{
		ID:           uuid.New(),
		AirportID:    airport.ID,
		Name:         adminName,
		Email:        email,
		NIP:          strings.TrimSpace(req.NIP),
		PasswordHash: hashedPassword,
		Role:         "airport_admin",
		Phone:        strings.TrimSpace(req.Phone),
		IsActive:     true,
		IsVerified:   false, // Requires OTP email verification
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := h.store.CreateUser(r.Context(), adminUser); err != nil {
		http.Error(w, `{"error":"Failed to create admin user"}`, http.StatusInternalServerError)
		return
	}

	// 4. Generate 6-digit OTP code (Valid for 10 minutes)
	otpCode := auth.GenerateNumericOTP()
	if err := h.store.SaveOTP(r.Context(), email, otpCode, 10*time.Minute); err != nil {
		http.Error(w, `{"error":"Failed to generate security OTP"}`, http.StatusInternalServerError)
		return
	}

	// 5. Send OTP Email asynchronously
	go func() {
		_ = h.mailer.SendOTPEmail(email, adminName, otpCode)
	}()

	resp := model.RegisterResponse{
		Message:          "Pendaftaran bandara berhasil. Silakan cek email Anda untuk memasukkan 6-digit kode OTP verifikasi.",
		Email:            email,
		ExpiresInMinutes: 10,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// VerifyOTP handles email verification via 6-digit code and activates the account
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	otp := strings.TrimSpace(req.OTP)

	if email == "" || otp == "" {
		http.Error(w, `{"error":"Email and OTP code are required"}`, http.StatusBadRequest)
		return
	}

	// 1. Verify & Consume OTP code
	if err := h.store.ValidateAndConsumeOTP(r.Context(), email, otp); err != nil {
		http.Error(w, `{"error":"Kode OTP tidak valid atau sudah kedaluwarsa"}`, http.StatusUnauthorized)
		return
	}

	// 2. Mark user as verified
	if err := h.store.MarkUserVerified(r.Context(), email); err != nil {
		http.Error(w, `{"error":"Gagal memverifikasi user"}`, http.StatusInternalServerError)
		return
	}

	// 3. Fetch user & airport
	user, err := h.store.GetUserByIdentifier(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"User tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	airport, err := h.store.GetAirportByID(r.Context(), user.AirportID)
	if err != nil {
		http.Error(w, `{"error":"Bandara tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	// 4. Issue authenticated JWT Token
	token, expiresAt, err := auth.GenerateToken(user.ID, airport.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"Gagal menerbitkan token otentikasi"}`, http.StatusInternalServerError)
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

// ResendOTP generates a new 6-digit OTP and resends to the given email
func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.ResendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := h.store.GetUserByIdentifier(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"Email tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	if user.IsVerified {
		http.Error(w, `{"error":"Akun email ini sudah terverifikasi sebelumnya. Silakan langsung login."}`, http.StatusBadRequest)
		return
	}

	// Generate and send new OTP
	newOTP := auth.GenerateNumericOTP()
	if err := h.store.SaveOTP(r.Context(), email, newOTP, 10*time.Minute); err != nil {
		http.Error(w, `{"error":"Gagal membuat kode OTP baru"}`, http.StatusInternalServerError)
		return
	}

	go func() {
		_ = h.mailer.SendOTPEmail(email, user.Name, newOTP)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "Kode OTP baru telah dikirimkan ke email Anda.",
		"email":              email,
		"expires_in_minutes": 10,
	})
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

	// Check if email has been verified
	if !user.IsVerified {
		http.Error(w, `{"error":"Email belum diverifikasi. Silakan masukkan kode OTP yang telah dikirim ke email Anda untuk mengaktifkan akun."}`, http.StatusForbidden)
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
