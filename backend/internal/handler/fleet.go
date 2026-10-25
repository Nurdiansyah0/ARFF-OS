package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/lop"
	"github.com/Nurdiansyah0/arff-os-backend/internal/middleware"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
	"github.com/Nurdiansyah0/arff-os-backend/internal/repository"
)

type FleetHandler struct {
	store *repository.Store
}

func NewFleetHandler(store *repository.Store) *FleetHandler {
	return &FleetHandler{store: store}
}

// GetVehicles lists all vehicles for the authenticated user's airport
func (h *FleetHandler) GetVehicles(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	vehicles, err := h.store.GetVehiclesByAirport(r.Context(), claims.AirportID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch vehicles"})
		return
	}
	if vehicles == nil {
		vehicles = []*model.Vehicle{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":    len(vehicles),
		"vehicles": vehicles,
	})
}

// CreateVehicle registers a new emergency response vehicle
func (h *FleetHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req model.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.CallSign == "" || req.VehicleType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "call_sign and vehicle_type are required"})
		return
	}

	status := req.Status
	if status == "" {
		status = model.VehicleStatusServiceable
	}

	now := time.Now().UTC()
	v := &model.Vehicle{
		ID:                 uuid.New(),
		AirportID:          claims.AirportID,
		CallSign:           strings.TrimSpace(req.CallSign),
		RegistrationNumber: strings.TrimSpace(req.RegistrationNumber),
		VehicleType:        req.VehicleType,
		BrandChassis:       strings.TrimSpace(req.BrandChassis),
		YearManufactured:   req.YearManufactured,
		WaterCapacity:      req.WaterCapacity,
		FoamCapacity:       req.FoamCapacity,
		DCPCapacity:        req.DCPCapacity,
		CO2Capacity:        req.CO2Capacity,
		PumpDischargeRate:  req.PumpDischargeRate,
		RoofTurretRate:     req.RoofTurretRate,
		BumperTurretRate:   req.BumperTurretRate,
		UnderTruckNozzles:  req.UnderTruckNozzles,
		Status:             status,
		CurrentWaterLevel:  req.CurrentWaterLevel,
		CurrentFoamLevel:   req.CurrentFoamLevel,
		CurrentDCPLevel:    req.CurrentDCPLevel,
		CurrentN2Pressure:  req.CurrentN2Pressure,
		OdometerKM:         req.OdometerKM,
		PumpHours:          req.PumpHours,
		Notes:              req.Notes,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Default agent levels to full capacity if left 0
	if v.CurrentWaterLevel == 0 && v.WaterCapacity > 0 {
		v.CurrentWaterLevel = v.WaterCapacity
	}
	if v.CurrentFoamLevel == 0 && v.FoamCapacity > 0 {
		v.CurrentFoamLevel = v.FoamCapacity
	}
	if v.CurrentDCPLevel == 0 && v.DCPCapacity > 0 {
		v.CurrentDCPLevel = v.DCPCapacity
	}

	if err := h.store.CreateVehicle(r.Context(), v); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create vehicle"})
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

// GetVehicle retrieves single vehicle by ID
func (h *FleetHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	idStr := r.PathValue("id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid vehicle ID format"})
		return
	}

	vehicle, err := h.store.GetVehicleByID(r.Context(), claims.AirportID, vehicleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Vehicle not found"})
		return
	}

	writeJSON(w, http.StatusOK, vehicle)
}

// UpdateVehicle updates a vehicle's specifications or readiness status
func (h *FleetHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	idStr := r.PathValue("id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid vehicle ID format"})
		return
	}

	vehicle, err := h.store.GetVehicleByID(r.Context(), claims.AirportID, vehicleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Vehicle not found"})
		return
	}

	var req model.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.CallSign != "" {
		vehicle.CallSign = strings.TrimSpace(req.CallSign)
	}
	if req.RegistrationNumber != "" {
		vehicle.RegistrationNumber = strings.TrimSpace(req.RegistrationNumber)
	}
	if req.Status != "" {
		vehicle.Status = req.Status
	}
	if req.WaterCapacity > 0 {
		vehicle.WaterCapacity = req.WaterCapacity
	}
	if req.FoamCapacity > 0 {
		vehicle.FoamCapacity = req.FoamCapacity
	}
	if req.DCPCapacity > 0 {
		vehicle.DCPCapacity = req.DCPCapacity
	}
	if req.PumpDischargeRate > 0 {
		vehicle.PumpDischargeRate = req.PumpDischargeRate
	}
	if req.CurrentWaterLevel >= 0 {
		vehicle.CurrentWaterLevel = req.CurrentWaterLevel
	}
	if req.CurrentFoamLevel >= 0 {
		vehicle.CurrentFoamLevel = req.CurrentFoamLevel
	}
	if req.CurrentDCPLevel >= 0 {
		vehicle.CurrentDCPLevel = req.CurrentDCPLevel
	}
	if req.CurrentN2Pressure > 0 {
		vehicle.CurrentN2Pressure = req.CurrentN2Pressure
	}
	if req.OdometerKM > 0 {
		vehicle.OdometerKM = req.OdometerKM
	}
	if req.PumpHours > 0 {
		vehicle.PumpHours = req.PumpHours
	}
	if req.Notes != "" {
		vehicle.Notes = req.Notes
	}

	if err := h.store.UpdateVehicle(r.Context(), vehicle); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update vehicle"})
		return
	}

	writeJSON(w, http.StatusOK, vehicle)
}

// GetLoPStatus computes dynamic ICAO Level of Protection for this airport
func (h *FleetHandler) GetLoPStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	airport, err := h.store.GetAirportByID(r.Context(), claims.AirportID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Airport tenant not found"})
		return
	}

	vehicles, err := h.store.GetVehiclesByAirport(r.Context(), claims.AirportID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch fleet data"})
		return
	}

	lopStatus := lop.CalculateLoP(airport.ID, airport.Name, airport.Category, vehicles)
	writeJSON(w, http.StatusOK, lopStatus)
}

// SubmitInspection records a daily vehicle checklist and updates vehicle levels
func (h *FleetHandler) SubmitInspection(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req model.VehicleInspection
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.VehicleID == uuid.Nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "vehicle_id is required"})
		return
	}

	vehicle, err := h.store.GetVehicleByID(r.Context(), claims.AirportID, req.VehicleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Vehicle not found in this airport"})
		return
	}

	now := time.Now().UTC()
	req.ID = uuid.New()
	req.AirportID = claims.AirportID
	req.InspectorID = claims.UserID
	req.CreatedAt = now

	if req.OverallReadiness == "" {
		req.OverallReadiness = model.VehicleStatusServiceable
	}

	if err := h.store.CreateVehicleInspection(r.Context(), &req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to record vehicle inspection"})
		return
	}

	// Synchronize vehicle's live operational telemetry and agent levels
	vehicle.CurrentWaterLevel = req.WaterLevelLiters
	vehicle.CurrentFoamLevel = req.FoamLevelLiters
	vehicle.CurrentDCPLevel = req.DCPLevelKg
	if req.OdometerKM > vehicle.OdometerKM {
		vehicle.OdometerKM = req.OdometerKM
	}
	if req.PumpHours > vehicle.PumpHours {
		vehicle.PumpHours = req.PumpHours
	}
	vehicle.Status = req.OverallReadiness
	_ = h.store.UpdateVehicle(r.Context(), vehicle)

	writeJSON(w, http.StatusCreated, req)
}

// GetInspections returns the inspection logs
func (h *FleetHandler) GetInspections(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var vehicleID *uuid.UUID
	vIDStr := r.URL.Query().Get("vehicle_id")
	if vIDStr != "" {
		if parsed, err := uuid.Parse(vIDStr); err == nil {
			vehicleID = &parsed
		}
	}

	inspections, err := h.store.GetVehicleInspections(r.Context(), claims.AirportID, vehicleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch inspection logs"})
		return
	}
	if inspections == nil {
		inspections = []*model.VehicleInspection{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":       len(inspections),
		"inspections": inspections,
	})
}

// GetStocks lists all extinguishing agent bulk stocks
func (h *FleetHandler) GetStocks(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	stocks, err := h.store.GetAgentStocksByAirport(r.Context(), claims.AirportID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch extinguishing agent stocks"})
		return
	}
	if stocks == nil {
		stocks = []*model.ExtinguishingAgentStock{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":  len(stocks),
		"stocks": stocks,
	})
}
