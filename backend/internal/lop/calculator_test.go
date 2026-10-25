package lop_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/lop"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

func TestCalculateLoP_OptimalCat9(t *testing.T) {
	airportID := uuid.New()
	airportName := "Soekarno-Hatta International (CGK)"
	declaredCat := 9

	// Cat 9 requires: Water >= 36,400 L, Discharge >= 13,500 L/min, DCP >= 450 kg, Major Vehicles >= 3
	vehicles := []*model.Vehicle{
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 01",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusServiceable,
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 02",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusServiceable,
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 03",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusServiceable,
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
	}

	res := lop.CalculateLoP(airportID, airportName, declaredCat, vehicles)

	if res.IsDegraded {
		t.Errorf("Expected compliance to be OPTIMAL, got degraded")
	}
	if res.RealizedCategory != 9 {
		t.Errorf("Expected RealizedCategory 9, got %d", res.RealizedCategory)
	}
	if res.ComplianceStatus != "OPTIMAL" {
		t.Errorf("Expected compliance status OPTIMAL, got %s", res.ComplianceStatus)
	}
	if res.AvailableWaterLiters != 37500 {
		t.Errorf("Expected 37500L available water, got %d", res.AvailableWaterLiters)
	}
}

func TestCalculateLoP_DegradedWhenMajorVehicleVOR(t *testing.T) {
	airportID := uuid.New()
	airportName := "Hang Nadim International (BTH)"
	declaredCat := 9

	// 1 Foam Tender breaks down (VOR)
	vehicles := []*model.Vehicle{
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 01",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusServiceable,
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 02",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusServiceable,
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
		{
			ID:                uuid.New(),
			AirportID:         airportID,
			CallSign:          "Foam Tender 03",
			VehicleType:       model.VehicleTypeFoamTender,
			Status:            model.VehicleStatusVOR, // BROKEN DOWN
			WaterCapacity:     12500,
			CurrentWaterLevel: 12500,
			DCPCapacity:       250,
			CurrentDCPLevel:   250,
			PumpDischargeRate: 6000,
		},
	}

	res := lop.CalculateLoP(airportID, airportName, declaredCat, vehicles)

	if !res.IsDegraded {
		t.Errorf("Expected status to be degraded when FT-03 is VOR")
	}
	if res.VORVehicles != 1 {
		t.Errorf("Expected 1 VOR vehicle, got %d", res.VORVehicles)
	}
	// With 25,000L water and 2 major vehicles, max category is Cat 7 (Cat 8 requires 27,300L and 3 vehicles)
	if res.RealizedCategory != 7 {
		t.Errorf("Expected degraded category to be 7, got %d", res.RealizedCategory)
	}
	if res.DraftNOTAM == "" {
		t.Errorf("Expected draft NOTAM to be generated for degraded state")
	}
}
