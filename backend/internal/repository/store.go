package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAirportNotFound = errors.New("airport not found")
	ErrEmailExists     = errors.New("email is already registered")
	ErrInvalidOTP      = errors.New("invalid or expired OTP code")
)

type otpRecord struct {
	Code      string
	ExpiresAt time.Time
}

type Store struct {
	db *sql.DB

	// In-memory mock storage for offline / local fallback mode
	mu       sync.RWMutex
	airports map[uuid.UUID]*model.Airport
	users    map[uuid.UUID]*model.User
	otps     map[string]*otpRecord // key: email
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:       db,
		airports: make(map[uuid.UUID]*model.Airport),
		users:    make(map[uuid.UUID]*model.User),
		otps:     make(map[string]*otpRecord),
	}
}

// CreateAirport inserts a new airport tenant
func (s *Store) CreateAirport(ctx context.Context, airport *model.Airport) error {
	if s.db != nil {
		query := `
			INSERT INTO airports (id, name, iata_code, icao_code, category, country, timezone, plan, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err := s.db.ExecContext(ctx, query,
			airport.ID, airport.Name, airport.IATACode, airport.ICAOCode,
			airport.Category, airport.Country, airport.Timezone, airport.Plan,
			airport.Status, airport.CreatedAt, airport.UpdatedAt,
		)
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.airports[airport.ID] = airport
	return nil
}

// GetAirportByID retrieves an airport by UUID
func (s *Store) GetAirportByID(ctx context.Context, id uuid.UUID) (*model.Airport, error) {
	if s.db != nil {
		query := `SELECT id, name, iata_code, icao_code, category, country, timezone, plan, status, created_at, updated_at FROM airports WHERE id = $1`
		row := s.db.QueryRowContext(ctx, query, id)
		var a model.Airport
		var iata, icao sql.NullString
		err := row.Scan(&a.ID, &a.Name, &iata, &icao, &a.Category, &a.Country, &a.Timezone, &a.Plan, &a.Status, &a.CreatedAt, &a.UpdatedAt)
		if err == nil {
			if iata.Valid {
				a.IATACode = iata.String
			}
			if icao.Valid {
				a.ICAOCode = icao.String
			}
			return &a, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.airports[id]
	if !ok {
		return nil, ErrAirportNotFound
	}
	return a, nil
}

// CreateUser inserts a new user
func (s *Store) CreateUser(ctx context.Context, user *model.User) error {
	if s.db != nil {
		query := `
			INSERT INTO users (id, airport_id, name, email, nip, password_hash, role, phone, is_active, is_verified, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err := s.db.ExecContext(ctx, query,
			user.ID, user.AirportID, user.Name, user.Email, user.NIP,
			user.PasswordHash, user.Role, user.Phone, user.IsActive, user.IsVerified,
			user.CreatedAt, user.UpdatedAt,
		)
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if strings.EqualFold(existing.Email, user.Email) {
			return ErrEmailExists
		}
	}
	s.users[user.ID] = user
	return nil
}

// GetUserByIdentifier retrieves user by either Email or NIP
func (s *Store) GetUserByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	if s.db != nil {
		query := `SELECT id, airport_id, name, email, nip, password_hash, role, phone, is_active, is_verified, created_at, updated_at FROM users WHERE email = $1 OR nip = $1 LIMIT 1`
		row := s.db.QueryRowContext(ctx, query, identifier)
		var u model.User
		var nip, phone sql.NullString
		err := row.Scan(&u.ID, &u.AirportID, &u.Name, &u.Email, &nip, &u.PasswordHash, &u.Role, &phone, &u.IsActive, &u.IsVerified, &u.CreatedAt, &u.UpdatedAt)
		if err == nil {
			if nip.Valid {
				u.NIP = nip.String
			}
			if phone.Valid {
				u.Phone = phone.String
			}
			return &u, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if strings.EqualFold(u.Email, identifier) || (u.NIP != "" && u.NIP == identifier) {
			return u, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetUserByID retrieves user by UUID
func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if s.db != nil {
		query := `SELECT id, airport_id, name, email, nip, password_hash, role, phone, is_active, is_verified, created_at, updated_at FROM users WHERE id = $1`
		row := s.db.QueryRowContext(ctx, query, id)
		var u model.User
		var nip, phone sql.NullString
		err := row.Scan(&u.ID, &u.AirportID, &u.Name, &u.Email, &nip, &u.PasswordHash, &u.Role, &phone, &u.IsActive, &u.IsVerified, &u.CreatedAt, &u.UpdatedAt)
		if err == nil {
			if nip.Valid {
				u.NIP = nip.String
			}
			if phone.Valid {
				u.Phone = phone.String
			}
			return &u, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// SaveOTP stores an OTP code with expiration
func (s *Store) SaveOTP(ctx context.Context, email, code string, duration time.Duration) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	expiresAt := time.Now().Add(duration)

	if s.db != nil {
		// Delete any existing OTP for this email
		_, _ = s.db.ExecContext(ctx, `DELETE FROM otp_verifications WHERE email = $1`, normalizedEmail)

		query := `INSERT INTO otp_verifications (id, email, code, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)`
		_, err := s.db.ExecContext(ctx, query, uuid.New(), normalizedEmail, code, expiresAt, time.Now())
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.otps[normalizedEmail] = &otpRecord{
		Code:      code,
		ExpiresAt: expiresAt,
	}
	return nil
}

// ValidateAndConsumeOTP verifies if OTP code matches and has not expired
func (s *Store) ValidateAndConsumeOTP(ctx context.Context, email, code string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	if s.db != nil {
		var storedCode string
		var expiresAt time.Time
		query := `SELECT code, expires_at FROM otp_verifications WHERE email = $1 ORDER BY created_at DESC LIMIT 1`
		err := s.db.QueryRowContext(ctx, query, normalizedEmail).Scan(&storedCode, &expiresAt)
		if err == nil {
			if time.Now().After(expiresAt) || storedCode != strings.TrimSpace(code) {
				return ErrInvalidOTP
			}
			// Delete consumed OTP
			_, _ = s.db.ExecContext(ctx, `DELETE FROM otp_verifications WHERE email = $1`, normalizedEmail)
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.otps[normalizedEmail]
	if !ok {
		return ErrInvalidOTP
	}

	if time.Now().After(rec.ExpiresAt) || rec.Code != strings.TrimSpace(code) {
		return ErrInvalidOTP
	}

	delete(s.otps, normalizedEmail)
	return nil
}

// MarkUserVerified sets is_verified to true for the given user email
func (s *Store) MarkUserVerified(ctx context.Context, email string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	if s.db != nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE users SET is_verified = TRUE, updated_at = NOW() WHERE email = $1`, normalizedEmail)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if strings.EqualFold(u.Email, normalizedEmail) {
			u.IsVerified = true
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return nil
}
