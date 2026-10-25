package model

import (
	"time"

	"github.com/google/uuid"
)

// Vehicle Types
const (
	VehicleTypeFoamTender = "foam_tender"
	VehicleTypeRIV        = "riv"
	VehicleTypeNurseTender = "nurse_tender"
	VehicleTypeAmbulance  = "ambulance"
	VehicleTypeCommandCar = "command_car"
	VehicleTypeRescueBoat = "rescue_boat"
)

// Vehicle Statuses
const (
	VehicleStatusServiceable = "serviceable" // Ready for immediate operational dispatch
	VehicleStatusLimited     = "limited"     // Minor defect, limited capacity
	VehicleStatusVOR         = "vor"         // Vehicle Off Road (Unserviceable / Breakdown)
)

// Vehicle represents an ARFF emergency vehicle tied to a specific airport tenant
type Vehicle struct {
	ID                 uuid.UUID `json:"id"`
	AirportID          uuid.UUID `json:"airport_id"`
	CallSign           string    `json:"call_sign"`
	RegistrationNumber string    `json:"registration_number"`
	VehicleType        string    `json:"vehicle_type"`
	BrandChassis       string    `json:"brand_chassis"`
	YearManufactured   int       `json:"year_manufactured"`

	// Extinguishing Agent Capacities
	WaterCapacity int `json:"water_capacity"` // Liters
	FoamCapacity  int `json:"foam_capacity"`  // Liters
	DCPCapacity   int `json:"dcp_capacity"`   // kg
	CO2Capacity   int `json:"co2_capacity"`   // kg

	// Discharge Rates (ICAO Annex 14)
	PumpDischargeRate int  `json:"pump_discharge_rate"` // L/min
	RoofTurretRate    int  `json:"roof_turret_rate"`    // L/min
	BumperTurretRate  int  `json:"bumper_turret_rate"`  // L/min
	UnderTruckNozzles bool `json:"under_truck_nozzles"`

	// Real-time Operational Status & Current Levels
	Status            string  `json:"status"` // 'serviceable', 'limited', 'vor'
	CurrentWaterLevel int     `json:"current_water_level"`
	CurrentFoamLevel  int     `json:"current_foam_level"`
	CurrentDCPLevel   int     `json:"current_dcp_level"`
	CurrentN2Pressure float64 `json:"current_n2_pressure"`

	// Telemetry
	OdometerKM float64 `json:"odometer_km"`
	PumpHours  float64 `json:"pump_hours"`
	Notes      string  `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ExtinguishingAgentStock represents airport bulk storage for fire-extinguishing agents
type ExtinguishingAgentStock struct {
	ID                      uuid.UUID  `json:"id"`
	AirportID               uuid.UUID  `json:"airport_id"`
	AgentType               string     `json:"agent_type"` // 'water', 'foam_afff_3', 'foam_afff_6', 'dcp_powder'
	StorageLocation         string     `json:"storage_location"`
	CurrentQuantity         float64    `json:"current_quantity"`
	MinimumReserveThreshold float64    `json:"minimum_reserve_threshold"` // ICAO 200% reserve
	Unit                    string     `json:"unit"`                      // 'liters', 'kg'
	BatchNumber             string     `json:"batch_number,omitempty"`
	ExpiryDate              *time.Time `json:"expiry_date,omitempty"`
	LastTestedAt            *time.Time `json:"last_tested_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// VehicleInspection represents a daily shift inspection & functional pump test
type VehicleInspection struct {
	ID             uuid.UUID `json:"id"`
	AirportID      uuid.UUID `json:"airport_id"`
	VehicleID      uuid.UUID `json:"vehicle_id"`
	InspectorID    uuid.UUID `json:"inspector_id"`
	InspectorName  string    `json:"inspector_name,omitempty"`
	Shift          string    `json:"shift"` // 'pagi', 'siang', 'malam'
	Regu           string    `json:"regu"`  // 'Regu A', 'Regu B', 'Regu C', 'Regu D'

	// Functional item checks ('pass', 'fail', 'warn')
	EngineStatus          string `json:"engine_status"`
	BrakeStatus           string `json:"brake_status"`
	TransmissionStatus    string `json:"transmission_status"`
	LightingSirenStatus   string `json:"lighting_siren_status"`
	PumpTestStatus        string `json:"pump_test_status"`
	TurretTestStatus      string `json:"turret_test_status"`
	UnderTruckTestStatus  string `json:"under_truck_test_status"`
	SCBAReadinessStatus   string `json:"scba_readiness_status"`
	ExtricationToolsStatus string `json:"extrication_tools_status"`

	// Metrics
	WaterLevelLiters  int     `json:"water_level_liters"`
	FoamLevelLiters   int     `json:"foam_level_liters"`
	DCPLevelKg        int     `json:"dcp_level_kg"`
	FuelLevelPercent  int     `json:"fuel_level_percent"`
	EngineOilPressure float64 `json:"engine_oil_pressure"`
	PumpPressureBar   float64 `json:"pump_pressure_bar"`
	OdometerKM        float64 `json:"odometer_km"`
	PumpHours         float64 `json:"pump_hours"`

	Notes            string    `json:"notes,omitempty"`
	OverallReadiness string    `json:"overall_readiness"` // 'serviceable', 'limited', 'vor'
	CreatedAt        time.Time `json:"created_at"`
}

// CreateVehicleRequest payload
type CreateVehicleRequest struct {
	CallSign           string  `json:"call_sign"`
	RegistrationNumber string  `json:"registration_number"`
	VehicleType        string  `json:"vehicle_type"`
	BrandChassis       string  `json:"brand_chassis"`
	YearManufactured   int     `json:"year_manufactured"`
	WaterCapacity      int     `json:"water_capacity"`
	FoamCapacity       int     `json:"foam_capacity"`
	DCPCapacity        int     `json:"dcp_capacity"`
	CO2Capacity        int     `json:"co2_capacity"`
	PumpDischargeRate  int     `json:"pump_discharge_rate"`
	RoofTurretRate     int     `json:"roof_turret_rate"`
	BumperTurretRate   int     `json:"bumper_turret_rate"`
	UnderTruckNozzles  bool    `json:"under_truck_nozzles"`
	Status             string  `json:"status"`
	CurrentWaterLevel  int     `json:"current_water_level"`
	CurrentFoamLevel   int     `json:"current_foam_level"`
	CurrentDCPLevel    int     `json:"current_dcp_level"`
	CurrentN2Pressure  float64 `json:"current_n2_pressure"`
	OdometerKM         float64 `json:"odometer_km"`
	PumpHours          float64 `json:"pump_hours"`
	Notes              string  `json:"notes"`
}

// LevelOfProtectionStatus represents dynamic ICAO Annex 14 compliance calculation
type LevelOfProtectionStatus struct {
	AirportID           uuid.UUID `json:"airport_id"`
	AirportName         string    `json:"airport_name"`
	DeclaredCategory    int       `json:"declared_category"`    // ICAO Target Category (e.g. Cat 9)
	RealizedCategory    int       `json:"realized_category"`    // Actual Available Category based on serviceable fleet
	IsDegraded          bool      `json:"is_degraded"`          // True if Realized < Declared
	ComplianceStatus    string    `json:"compliance_status"`    // 'OPTIMAL', 'DEGRADED', 'CRITICAL'
	
	// Fleet Numbers
	TotalVehicles       int `json:"total_vehicles"`
	ServiceableVehicles int `json:"serviceable_vehicles"`
	VORVehicles         int `json:"vor_vehicles"`
	
	// Agent Capacities & Discharge Comparison
	RequiredWaterLiters    int `json:"required_water_liters"`
	AvailableWaterLiters   int `json:"available_water_liters"`
	RequiredDischargeRate  int `json:"required_discharge_rate"`  // L/min
	AvailableDischargeRate int `json:"available_discharge_rate"` // L/min
	RequiredDCPKg          int `json:"required_dcp_kg"`
	AvailableDCPKg         int `json:"available_dcp_kg"`
	RequiredVehiclesCount  int `json:"required_vehicles_count"`
	AvailableMajorVehicles int `json:"available_major_vehicles"`

	// Notification & NOTAM Advisory
	AlertMessage string `json:"alert_message,omitempty"`
	DraftNOTAM   string `json:"draft_notam,omitempty"`
}
