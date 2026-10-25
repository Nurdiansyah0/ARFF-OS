package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAirportNotFound = errors.New("airport not found")
	ErrEmailExists     = errors.New("email is already registered")
)

type Store struct {
	db *sql.DB

	// In-memory mock storage for offline / local fallback mode
	mu       sync.RWMutex
	airports map[uuid.UUID]*model.Airport
	users    map[uuid.UUID]*model.User
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:       db,
		airports: make(map[uuid.UUID]*model.Airport),
		users:    make(map[uuid.UUID]*model.User),
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
		// If SQL execution failed (e.g. table not created yet), fallback to memory in dev
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
			INSERT INTO users (id, airport_id, name, email, nip, password_hash, role, phone, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err := s.db.ExecContext(ctx, query,
			user.ID, user.AirportID, user.Name, user.Email, user.NIP,
			user.PasswordHash, user.Role, user.Phone, user.IsActive,
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
		query := `SELECT id, airport_id, name, email, nip, password_hash, role, phone, is_active, created_at, updated_at FROM users WHERE email = $1 OR nip = $1 LIMIT 1`
		row := s.db.QueryRowContext(ctx, query, identifier)
		var u model.User
		var nip, phone sql.NullString
		err := row.Scan(&u.ID, &u.AirportID, &u.Name, &u.Email, &nip, &u.PasswordHash, &u.Role, &phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
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
		query := `SELECT id, airport_id, name, email, nip, password_hash, role, phone, is_active, created_at, updated_at FROM users WHERE id = $1`
		row := s.db.QueryRowContext(ctx, query, id)
		var u model.User
		var nip, phone sql.NullString
		err := row.Scan(&u.ID, &u.AirportID, &u.Name, &u.Email, &nip, &u.PasswordHash, &u.Role, &phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
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
