package model

import (
	"time"

	"github.com/google/uuid"
)

type Airport struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IATACode  string    `json:"iata_code,omitempty"`
	ICAOCode  string    `json:"icao_code,omitempty"`
	Category  int       `json:"category"`
	Country   string    `json:"country"`
	Timezone  string    `json:"timezone"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterAirportRequest struct {
	// Airport Details
	AirportName string `json:"airport_name"`
	IATACode    string `json:"iata_code"`
	ICAOCode    string `json:"icao_code"`
	Category    int    `json:"category"`
	Country     string `json:"country"`

	// Initial Admin Details
	AdminName string `json:"admin_name"`
	Email     string `json:"email"`
	NIP       string `json:"nip"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
}
