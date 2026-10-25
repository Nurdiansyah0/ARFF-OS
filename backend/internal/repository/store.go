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
	ErrUserNotFound     = errors.New("user not found")
	ErrAirportNotFound  = errors.New("airport not found")
	ErrEmailExists      = errors.New("email is already registered")
	ErrInvalidOTP       = errors.New("invalid or expired OTP code")
	ErrVehicleNotFound  = errors.New("vehicle not found")
	ErrApprovalNotFound = errors.New("approval not found")
	ErrStockNotFound    = errors.New("extinguishing agent stock not found")
)

type otpRecord struct {
	Code      string
	ExpiresAt time.Time
}

type Store struct {
	db *sql.DB

	// In-memory mock storage for offline / local fallback mode
	mu          sync.RWMutex
	airports    map[uuid.UUID]*model.Airport
	users       map[uuid.UUID]*model.User
	otps        map[string]*otpRecord // key: email
	vehicles    map[uuid.UUID]*model.Vehicle
	stocks      map[uuid.UUID]*model.ExtinguishingAgentStock
	inspections map[uuid.UUID]*model.VehicleInspection
	approvals   map[uuid.UUID]*model.Approval
	actions     map[uuid.UUID][]*model.ApprovalAction // key: approval_id
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:          db,
		airports:    make(map[uuid.UUID]*model.Airport),
		users:       make(map[uuid.UUID]*model.User),
		otps:        make(map[string]*otpRecord),
		vehicles:    make(map[uuid.UUID]*model.Vehicle),
		stocks:      make(map[uuid.UUID]*model.ExtinguishingAgentStock),
		inspections: make(map[uuid.UUID]*model.VehicleInspection),
		approvals:   make(map[uuid.UUID]*model.Approval),
		actions:     make(map[uuid.UUID][]*model.ApprovalAction),
	}
}

// -------------------------------------------------------------
// Airport Tenant Methods
// -------------------------------------------------------------

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

// -------------------------------------------------------------
// User Methods
// -------------------------------------------------------------

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

// -------------------------------------------------------------
// OTP Methods
// -------------------------------------------------------------

func (s *Store) SaveOTP(ctx context.Context, email, code string, duration time.Duration) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	expiresAt := time.Now().Add(duration)

	if s.db != nil {
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

// -------------------------------------------------------------
// Vehicles (Fleet) Methods - Tenant Isolated by airport_id
// -------------------------------------------------------------

func (s *Store) CreateVehicle(ctx context.Context, v *model.Vehicle) error {
	if s.db != nil {
		query := `
			INSERT INTO vehicles (
				id, airport_id, call_sign, registration_number, vehicle_type, brand_chassis, year_manufactured,
				water_capacity, foam_capacity, dcp_capacity, co2_capacity,
				pump_discharge_rate, roof_turret_rate, bumper_turret_rate, under_truck_nozzles,
				status, current_water_level, current_foam_level, current_dcp_level, current_n2_pressure,
				odometer_km, pump_hours, notes, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, $11,
				$12, $13, $14, $15,
				$16, $17, $18, $19, $20,
				$21, $22, $23, $24, $25
			)
		`
		_, err := s.db.ExecContext(ctx, query,
			v.ID, v.AirportID, v.CallSign, v.RegistrationNumber, v.VehicleType, v.BrandChassis, v.YearManufactured,
			v.WaterCapacity, v.FoamCapacity, v.DCPCapacity, v.CO2Capacity,
			v.PumpDischargeRate, v.RoofTurretRate, v.BumperTurretRate, v.UnderTruckNozzles,
			v.Status, v.CurrentWaterLevel, v.CurrentFoamLevel, v.CurrentDCPLevel, v.CurrentN2Pressure,
			v.OdometerKM, v.PumpHours, v.Notes, v.CreatedAt, v.UpdatedAt,
		)
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.vehicles[v.ID] = v
	return nil
}

func (s *Store) GetVehiclesByAirport(ctx context.Context, airportID uuid.UUID) ([]*model.Vehicle, error) {
	if s.db != nil {
		query := `
			SELECT id, airport_id, call_sign, registration_number, vehicle_type, brand_chassis, year_manufactured,
			       water_capacity, foam_capacity, dcp_capacity, co2_capacity,
			       pump_discharge_rate, roof_turret_rate, bumper_turret_rate, under_truck_nozzles,
			       status, current_water_level, current_foam_level, current_dcp_level, current_n2_pressure,
			       odometer_km, pump_hours, notes, created_at, updated_at
			FROM vehicles WHERE airport_id = $1 ORDER BY call_sign ASC
		`
		rows, err := s.db.QueryContext(ctx, query, airportID)
		if err == nil {
			defer rows.Close()
			var list []*model.Vehicle
			for rows.Next() {
				var v model.Vehicle
				var reg, brand, notes sql.NullString
				err := rows.Scan(
					&v.ID, &v.AirportID, &v.CallSign, &reg, &v.VehicleType, &brand, &v.YearManufactured,
					&v.WaterCapacity, &v.FoamCapacity, &v.DCPCapacity, &v.CO2Capacity,
					&v.PumpDischargeRate, &v.RoofTurretRate, &v.BumperTurretRate, &v.UnderTruckNozzles,
					&v.Status, &v.CurrentWaterLevel, &v.CurrentFoamLevel, &v.CurrentDCPLevel, &v.CurrentN2Pressure,
					&v.OdometerKM, &v.PumpHours, &notes, &v.CreatedAt, &v.UpdatedAt,
				)
				if err == nil {
					if reg.Valid {
						v.RegistrationNumber = reg.String
					}
					if brand.Valid {
						v.BrandChassis = brand.String
					}
					if notes.Valid {
						v.Notes = notes.String
					}
					list = append(list, &v)
				}
			}
			return list, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*model.Vehicle
	for _, v := range s.vehicles {
		if v.AirportID == airportID {
			list = append(list, v)
		}
	}
	return list, nil
}

func (s *Store) GetVehicleByID(ctx context.Context, airportID uuid.UUID, vehicleID uuid.UUID) (*model.Vehicle, error) {
	if s.db != nil {
		query := `
			SELECT id, airport_id, call_sign, registration_number, vehicle_type, brand_chassis, year_manufactured,
			       water_capacity, foam_capacity, dcp_capacity, co2_capacity,
			       pump_discharge_rate, roof_turret_rate, bumper_turret_rate, under_truck_nozzles,
			       status, current_water_level, current_foam_level, current_dcp_level, current_n2_pressure,
			       odometer_km, pump_hours, notes, created_at, updated_at
			FROM vehicles WHERE id = $1 AND airport_id = $2
		`
		row := s.db.QueryRowContext(ctx, query, vehicleID, airportID)
		var v model.Vehicle
		var reg, brand, notes sql.NullString
		err := row.Scan(
			&v.ID, &v.AirportID, &v.CallSign, &reg, &v.VehicleType, &brand, &v.YearManufactured,
			&v.WaterCapacity, &v.FoamCapacity, &v.DCPCapacity, &v.CO2Capacity,
			&v.PumpDischargeRate, &v.RoofTurretRate, &v.BumperTurretRate, &v.UnderTruckNozzles,
			&v.Status, &v.CurrentWaterLevel, &v.CurrentFoamLevel, &v.CurrentDCPLevel, &v.CurrentN2Pressure,
			&v.OdometerKM, &v.PumpHours, &notes, &v.CreatedAt, &v.UpdatedAt,
		)
		if err == nil {
			if reg.Valid {
				v.RegistrationNumber = reg.String
			}
			if brand.Valid {
				v.BrandChassis = brand.String
			}
			if notes.Valid {
				v.Notes = notes.String
			}
			return &v, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.vehicles[vehicleID]
	if !ok || v.AirportID != airportID {
		return nil, ErrVehicleNotFound
	}
	return v, nil
}

func (s *Store) UpdateVehicle(ctx context.Context, v *model.Vehicle) error {
	v.UpdatedAt = time.Now().UTC()

	if s.db != nil {
		query := `
			UPDATE vehicles SET
				call_sign = $1, registration_number = $2, vehicle_type = $3, brand_chassis = $4, year_manufactured = $5,
				water_capacity = $6, foam_capacity = $7, dcp_capacity = $8, co2_capacity = $9,
				pump_discharge_rate = $10, roof_turret_rate = $11, bumper_turret_rate = $12, under_truck_nozzles = $13,
				status = $14, current_water_level = $15, current_foam_level = $16, current_dcp_level = $17, current_n2_pressure = $18,
				odometer_km = $19, pump_hours = $20, notes = $21, updated_at = $22
			WHERE id = $23 AND airport_id = $24
		`
		_, err := s.db.ExecContext(ctx, query,
			v.CallSign, v.RegistrationNumber, v.VehicleType, v.BrandChassis, v.YearManufactured,
			v.WaterCapacity, v.FoamCapacity, v.DCPCapacity, v.CO2Capacity,
			v.PumpDischargeRate, v.RoofTurretRate, v.BumperTurretRate, v.UnderTruckNozzles,
			v.Status, v.CurrentWaterLevel, v.CurrentFoamLevel, v.CurrentDCPLevel, v.CurrentN2Pressure,
			v.OdometerKM, v.PumpHours, v.Notes, v.UpdatedAt,
			v.ID, v.AirportID,
		)
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.vehicles[v.ID]
	if !ok || existing.AirportID != v.AirportID {
		return ErrVehicleNotFound
	}
	s.vehicles[v.ID] = v
	return nil
}

// -------------------------------------------------------------
// Vehicle Inspections
// -------------------------------------------------------------

func (s *Store) CreateVehicleInspection(ctx context.Context, insp *model.VehicleInspection) error {
	if s.db != nil {
		query := `
			INSERT INTO vehicle_inspections (
				id, airport_id, vehicle_id, inspector_id, shift, regu,
				engine_status, brake_status, transmission_status, lighting_siren_status,
				pump_test_status, turret_test_status, under_truck_test_status, scba_readiness_status, extrication_tools_status,
				water_level_liters, foam_level_liters, dcp_level_kg, fuel_level_percent,
				engine_oil_pressure, pump_pressure_bar, odometer_km, pump_hours, notes, overall_readiness, created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10,
				$11, $12, $13, $14, $15,
				$16, $17, $18, $19,
				$20, $21, $22, $23, $24, $25, $26
			)
		`
		_, err := s.db.ExecContext(ctx, query,
			insp.ID, insp.AirportID, insp.VehicleID, insp.InspectorID, insp.Shift, insp.Regu,
			insp.EngineStatus, insp.BrakeStatus, insp.TransmissionStatus, insp.LightingSirenStatus,
			insp.PumpTestStatus, insp.TurretTestStatus, insp.UnderTruckTestStatus, insp.SCBAReadinessStatus, insp.ExtricationToolsStatus,
			insp.WaterLevelLiters, insp.FoamLevelLiters, insp.DCPLevelKg, insp.FuelLevelPercent,
			insp.EngineOilPressure, insp.PumpPressureBar, insp.OdometerKM, insp.PumpHours, insp.Notes, insp.OverallReadiness, insp.CreatedAt,
		)
		if err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.inspections[insp.ID] = insp
	return nil
}

func (s *Store) GetVehicleInspections(ctx context.Context, airportID uuid.UUID, vehicleID *uuid.UUID) ([]*model.VehicleInspection, error) {
	if s.db != nil {
		query := `
			SELECT vi.id, vi.airport_id, vi.vehicle_id, vi.inspector_id, u.name, vi.shift, vi.regu,
			       vi.engine_status, vi.brake_status, vi.transmission_status, vi.lighting_siren_status,
			       vi.pump_test_status, vi.turret_test_status, vi.under_truck_test_status, vi.scba_readiness_status, vi.extrication_tools_status,
			       vi.water_level_liters, vi.foam_level_liters, vi.dcp_level_kg, vi.fuel_level_percent,
			       vi.engine_oil_pressure, vi.pump_pressure_bar, vi.odometer_km, vi.pump_hours, vi.notes, vi.overall_readiness, vi.created_at
			FROM vehicle_inspections vi
			LEFT JOIN users u ON vi.inspector_id = u.id
			WHERE vi.airport_id = $1
		`
		var args []interface{}
		args = append(args, airportID)
		if vehicleID != nil {
			query += " AND vi.vehicle_id = $2 ORDER BY vi.created_at DESC LIMIT 50"
			args = append(args, *vehicleID)
		} else {
			query += " ORDER BY vi.created_at DESC LIMIT 50"
		}

		rows, err := s.db.QueryContext(ctx, query, args...)
		if err == nil {
			defer rows.Close()
			var list []*model.VehicleInspection
			for rows.Next() {
				var vi model.VehicleInspection
				var name, notes sql.NullString
				err := rows.Scan(
					&vi.ID, &vi.AirportID, &vi.VehicleID, &vi.InspectorID, &name, &vi.Shift, &vi.Regu,
					&vi.EngineStatus, &vi.BrakeStatus, &vi.TransmissionStatus, &vi.LightingSirenStatus,
					&vi.PumpTestStatus, &vi.TurretTestStatus, &vi.UnderTruckTestStatus, &vi.SCBAReadinessStatus, &vi.ExtricationToolsStatus,
					&vi.WaterLevelLiters, &vi.FoamLevelLiters, &vi.DCPLevelKg, &vi.FuelLevelPercent,
					&vi.EngineOilPressure, &vi.PumpPressureBar, &vi.OdometerKM, &vi.PumpHours, &notes, &vi.OverallReadiness, &vi.CreatedAt,
				)
				if err == nil {
					if name.Valid {
						vi.InspectorName = name.String
					}
					if notes.Valid {
						vi.Notes = notes.String
					}
					list = append(list, &vi)
				}
			}
			return list, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*model.VehicleInspection
	for _, insp := range s.inspections {
		if insp.AirportID == airportID {
			if vehicleID == nil || insp.VehicleID == *vehicleID {
				list = append(list, insp)
			}
		}
	}
	return list, nil
}

// -------------------------------------------------------------
// Extinguishing Agent Stock Methods
// -------------------------------------------------------------

func (s *Store) GetAgentStocksByAirport(ctx context.Context, airportID uuid.UUID) ([]*model.ExtinguishingAgentStock, error) {
	if s.db != nil {
		query := `
			SELECT id, airport_id, agent_type, storage_location, current_quantity, minimum_reserve_threshold,
			       unit, batch_number, expiry_date, last_tested_at, created_at, updated_at
			FROM extinguishing_agent_stocks WHERE airport_id = $1 ORDER BY agent_type ASC
		`
		rows, err := s.db.QueryContext(ctx, query, airportID)
		if err == nil {
			defer rows.Close()
			var list []*model.ExtinguishingAgentStock
			for rows.Next() {
				var st model.ExtinguishingAgentStock
				var batch sql.NullString
				err := rows.Scan(
					&st.ID, &st.AirportID, &st.AgentType, &st.StorageLocation, &st.CurrentQuantity, &st.MinimumReserveThreshold,
					&st.Unit, &batch, &st.ExpiryDate, &st.LastTestedAt, &st.CreatedAt, &st.UpdatedAt,
				)
				if err == nil {
					if batch.Valid {
						st.BatchNumber = batch.String
					}
					list = append(list, &st)
				}
			}
			return list, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*model.ExtinguishingAgentStock
	for _, st := range s.stocks {
		if st.AirportID == airportID {
			list = append(list, st)
		}
	}
	return list, nil
}

// -------------------------------------------------------------
// Multi-Tier Approvals & Cryptographic Audit Trail
// -------------------------------------------------------------

func (s *Store) CreateApproval(ctx context.Context, app *model.Approval, initialAction *model.ApprovalAction) error {
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err == nil {
			queryApp := `
				INSERT INTO approvals (
					id, airport_id, entity_type, entity_id, title, description,
					current_stage, status, submitted_by, current_assignee_role,
					document_payload_hash, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`
			_, err = tx.ExecContext(ctx, queryApp,
				app.ID, app.AirportID, app.EntityType, app.EntityID, app.Title, app.Description,
				app.CurrentStage, app.Status, app.SubmittedBy, app.CurrentAssigneeRole,
				app.DocumentPayloadHash, app.CreatedAt, app.UpdatedAt,
			)
			if err == nil && initialAction != nil {
				queryAct := `
					INSERT INTO approval_actions (
						id, approval_id, actor_id, actor_name, actor_role, action,
						stage_from, stage_to, notes, signature_token, ip_address, user_agent, created_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				`
				_, err = tx.ExecContext(ctx, queryAct,
					initialAction.ID, initialAction.ApprovalID, initialAction.ActorID, initialAction.ActorName,
					initialAction.ActorRole, initialAction.Action, initialAction.StageFrom, initialAction.StageTo,
					initialAction.Notes, initialAction.SignatureToken, initialAction.IPAddress, initialAction.UserAgent, initialAction.CreatedAt,
				)
			}

			if err == nil {
				_ = tx.Commit()
				return nil
			}
			_ = tx.Rollback()
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[app.ID] = app
	if initialAction != nil {
		s.actions[app.ID] = append(s.actions[app.ID], initialAction)
	}
	return nil
}

func (s *Store) GetApprovalByID(ctx context.Context, airportID uuid.UUID, approvalID uuid.UUID) (*model.Approval, error) {
	if s.db != nil {
		query := `
			SELECT a.id, a.airport_id, a.entity_type, a.entity_id, a.title, a.description,
			       a.current_stage, a.status, a.submitted_by, u.name, a.current_assignee_role,
			       a.document_payload_hash, a.final_signature_hash, a.signed_at, a.created_at, a.updated_at
			FROM approvals a
			LEFT JOIN users u ON a.submitted_by = u.id
			WHERE a.id = $1 AND a.airport_id = $2
		`
		row := s.db.QueryRowContext(ctx, query, approvalID, airportID)
		var app model.Approval
		var desc, submitterName, finalSig sql.NullString
		err := row.Scan(
			&app.ID, &app.AirportID, &app.EntityType, &app.EntityID, &app.Title, &desc,
			&app.CurrentStage, &app.Status, &app.SubmittedBy, &submitterName, &app.CurrentAssigneeRole,
			&app.DocumentPayloadHash, &finalSig, &app.SignedAt, &app.CreatedAt, &app.UpdatedAt,
		)
		if err == nil {
			if desc.Valid {
				app.Description = desc.String
			}
			if submitterName.Valid {
				app.SubmitterName = submitterName.String
			}
			if finalSig.Valid {
				app.FinalSignatureHash = finalSig.String
			}

			// Fetch actions
			actQuery := `
				SELECT id, approval_id, actor_id, actor_name, actor_role, action,
				       stage_from, stage_to, notes, signature_token, ip_address, user_agent, created_at
				FROM approval_actions WHERE approval_id = $1 ORDER BY created_at ASC
			`
			actRows, err := s.db.QueryContext(ctx, actQuery, approvalID)
			if err == nil {
				defer actRows.Close()
				for actRows.Next() {
					var act model.ApprovalAction
					var notes, ip, ua sql.NullString
					_ = actRows.Scan(
						&act.ID, &act.ApprovalID, &act.ActorID, &act.ActorName, &act.ActorRole, &act.Action,
						&act.StageFrom, &act.StageTo, &notes, &act.SignatureToken, &ip, &ua, &act.CreatedAt,
					)
					if notes.Valid {
						act.Notes = notes.String
					}
					if ip.Valid {
						act.IPAddress = ip.String
					}
					if ua.Valid {
						act.UserAgent = ua.String
					}
					app.Actions = append(app.Actions, &act)
				}
			}
			return &app, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.approvals[approvalID]
	if !ok || app.AirportID != airportID {
		return nil, ErrApprovalNotFound
	}
	app.Actions = s.actions[approvalID]
	return app, nil
}

func (s *Store) GetApprovalsByAirport(ctx context.Context, airportID uuid.UUID, status string) ([]*model.Approval, error) {
	if s.db != nil {
		query := `
			SELECT a.id, a.airport_id, a.entity_type, a.entity_id, a.title, a.description,
			       a.current_stage, a.status, a.submitted_by, u.name, a.current_assignee_role,
			       a.document_payload_hash, a.final_signature_hash, a.signed_at, a.created_at, a.updated_at
			FROM approvals a
			LEFT JOIN users u ON a.submitted_by = u.id
			WHERE a.airport_id = $1
		`
		var args []interface{}
		args = append(args, airportID)
		if status != "" {
			query += " AND a.status = $2"
			args = append(args, status)
		}
		query += " ORDER BY a.created_at DESC LIMIT 50"

		rows, err := s.db.QueryContext(ctx, query, args...)
		if err == nil {
			defer rows.Close()
			var list []*model.Approval
			for rows.Next() {
				var app model.Approval
				var desc, submitterName, finalSig sql.NullString
				err := rows.Scan(
					&app.ID, &app.AirportID, &app.EntityType, &app.EntityID, &app.Title, &desc,
					&app.CurrentStage, &app.Status, &app.SubmittedBy, &submitterName, &app.CurrentAssigneeRole,
					&app.DocumentPayloadHash, &finalSig, &app.SignedAt, &app.CreatedAt, &app.UpdatedAt,
				)
				if err == nil {
					if desc.Valid {
						app.Description = desc.String
					}
					if submitterName.Valid {
						app.SubmitterName = submitterName.String
					}
					if finalSig.Valid {
						app.FinalSignatureHash = finalSig.String
					}
					list = append(list, &app)
				}
			}
			return list, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*model.Approval
	for _, app := range s.approvals {
		if app.AirportID == airportID {
			if status == "" || app.Status == status {
				list = append(list, app)
			}
		}
	}
	return list, nil
}

func (s *Store) SaveApprovalAction(ctx context.Context, app *model.Approval, action *model.ApprovalAction) error {
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err == nil {
			queryApp := `
				UPDATE approvals SET
					current_stage = $1, status = $2, current_assignee_role = $3,
					final_signature_hash = $4, signed_at = $5, updated_at = $6
				WHERE id = $7 AND airport_id = $8
			`
			_, err = tx.ExecContext(ctx, queryApp,
				app.CurrentStage, app.Status, app.CurrentAssigneeRole,
				app.FinalSignatureHash, app.SignedAt, app.UpdatedAt,
				app.ID, app.AirportID,
			)
			if err == nil {
				queryAct := `
					INSERT INTO approval_actions (
						id, approval_id, actor_id, actor_name, actor_role, action,
						stage_from, stage_to, notes, signature_token, ip_address, user_agent, created_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				`
				_, err = tx.ExecContext(ctx, queryAct,
					action.ID, action.ApprovalID, action.ActorID, action.ActorName,
					action.ActorRole, action.Action, action.StageFrom, action.StageTo,
					action.Notes, action.SignatureToken, action.IPAddress, action.UserAgent, action.CreatedAt,
				)
			}

			if err == nil {
				_ = tx.Commit()
				return nil
			}
			_ = tx.Rollback()
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[app.ID] = app
	s.actions[app.ID] = append(s.actions[app.ID], action)
	return nil
}
