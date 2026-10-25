package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type DashboardStats struct {
	FleetReadiness struct {
		Total   int `json:"total"`
		Ready   int `json:"ready"`
		VOR     int `json:"vor"`
		Unknown int `json:"unknown"`
	} `json:"fleet_readiness"`
	PersonnelStatus struct {
		OnDuty  int `json:"on_duty"`
		Standby int `json:"standby"`
		OffDuty int `json:"off_duty"`
	} `json:"personnel_status"`
	FlightActivity struct {
		TotalFlights int `json:"total_flights"`
		Arrivals     int `json:"arrivals"`
		Departures   int `json:"departures"`
	} `json:"flight_activity"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := DashboardStats{
		UpdatedAt: time.Now().UTC(),
	}

	// Sample starter data reflecting ARFF-OS operational requirements
	stats.FleetReadiness.Total = 12
	stats.FleetReadiness.Ready = 10
	stats.FleetReadiness.VOR = 2
	stats.FleetReadiness.Unknown = 0

	stats.PersonnelStatus.OnDuty = 24
	stats.PersonnelStatus.Standby = 8
	stats.PersonnelStatus.OffDuty = 16

	stats.FlightActivity.TotalFlights = 48
	stats.FlightActivity.Arrivals = 25
	stats.FlightActivity.Departures = 23

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}
