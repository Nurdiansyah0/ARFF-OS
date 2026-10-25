package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type HealthHandler struct {
	DBCheck func() bool
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbStatus := "disconnected"
	if h.DBCheck != nil && h.DBCheck() {
		dbStatus = "connected"
	}

	resp := HealthResponse{
		Status:    "healthy",
		Service:   "arff-os-backend",
		Version:   "v1.0.0",
		Database:  dbStatus,
		Timestamp: time.Now().UTC(),
		Message:   "ARFF-OS Backend Service Operational",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
