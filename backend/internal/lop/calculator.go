package lop

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

// ICAOCategorySpec holds minimum standard values required by ICAO Annex 14 Chapter 9
type ICAOCategorySpec struct {
	Category      int
	MinWaterL     int // Minimum usable water quantity (Liters)
	DischargeRate int // Minimum discharge rate (Liters / Minute)
	MinDCPKg      int // Minimum Dry Chemical Powder quantity (kg)
	MinVehicles   int // Minimum number of major operational rescue and firefighting vehicles
}

// ICAO Annex 14 Benchmark Table (Categories 1 to 10)
var ICAOStandards = map[int]ICAOCategorySpec{
	1:  {Category: 1, MinWaterL: 350, DischargeRate: 350, MinDCPKg: 45, MinVehicles: 1},
	2:  {Category: 2, MinWaterL: 1000, DischargeRate: 800, MinDCPKg: 90, MinVehicles: 1},
	3:  {Category: 3, MinWaterL: 1800, DischargeRate: 1300, MinDCPKg: 135, MinVehicles: 1},
	4:  {Category: 4, MinWaterL: 3600, DischargeRate: 2600, MinDCPKg: 135, MinVehicles: 1},
	5:  {Category: 5, MinWaterL: 8100, DischargeRate: 4500, MinDCPKg: 180, MinVehicles: 1},
	6:  {Category: 6, MinWaterL: 10700, DischargeRate: 6000, MinDCPKg: 225, MinVehicles: 2},
	7:  {Category: 7, MinWaterL: 15300, DischargeRate: 7500, MinDCPKg: 225, MinVehicles: 2},
	8:  {Category: 8, MinWaterL: 27300, DischargeRate: 10800, MinDCPKg: 450, MinVehicles: 3},
	9:  {Category: 9, MinWaterL: 36400, DischargeRate: 13500, MinDCPKg: 450, MinVehicles: 3},
	10: {Category: 10, MinWaterL: 48200, DischargeRate: 16600, MinDCPKg: 450, MinVehicles: 3},
}

// CalculateLoP evaluates the airport's fleet against ICAO Annex 14 Table 9-1 & 9-2
func CalculateLoP(airportID uuid.UUID, airportName string, declaredCategory int, vehicles []*model.Vehicle) *model.LevelOfProtectionStatus {
	if declaredCategory < 1 {
		declaredCategory = 1
	}
	if declaredCategory > 10 {
		declaredCategory = 10
	}

	totalVehicles := len(vehicles)
	serviceableVehicles := 0
	vorVehicles := 0

	availWater := 0
	availDischarge := 0
	availDCP := 0
	availMajorVehicles := 0

	for _, v := range vehicles {
		if v.Status == model.VehicleStatusServiceable {
			serviceableVehicles++

			// Check if major firefighting vehicle (Foam Tender or RIV)
			if v.VehicleType == model.VehicleTypeFoamTender || v.VehicleType == model.VehicleTypeRIV {
				availMajorVehicles++
			}

			// Tally usable extinguishing agents currently loaded on ready vehicles
			water := v.CurrentWaterLevel
			if water <= 0 {
				water = v.WaterCapacity
			}
			availWater += water

			dcp := v.CurrentDCPLevel
			if dcp <= 0 {
				dcp = v.DCPCapacity
			}
			availDCP += dcp

			// Tally discharge rate
			rate := v.PumpDischargeRate
			if rate <= 0 && v.RoofTurretRate > 0 {
				rate = v.RoofTurretRate + v.BumperTurretRate
			}
			availDischarge += rate
		} else if v.Status == model.VehicleStatusVOR {
			vorVehicles++
		}
	}

	// Calculate realized category by testing standards from 10 down to 1
	realizedCategory := 0
	for cat := 10; cat >= 1; cat-- {
		spec := ICAOStandards[cat]
		if availWater >= spec.MinWaterL &&
			availDischarge >= spec.DischargeRate &&
			availDCP >= spec.MinDCPKg &&
			availMajorVehicles >= spec.MinVehicles {
			realizedCategory = cat
			break
		}
	}

	specDeclared := ICAOStandards[declaredCategory]
	isDegraded := realizedCategory < declaredCategory

	complianceStatus := "OPTIMAL"
	var alertMsg string
	var draftNOTAM string

	if isDegraded {
		if realizedCategory == 0 {
			complianceStatus = "CRITICAL"
			alertMsg = fmt.Sprintf("CRITICAL ALERT: Airport %s has NO operational ARFF category capability. All major vehicles unserviceable or depleted.", airportName)
		} else {
			complianceStatus = "DEGRADED"
			alertMsg = fmt.Sprintf("WARNING: ARFF Category degraded from Cat %d to Cat %d due to unserviceable fleet (VOR: %d vehicles). Immediate NOTAM required.", declaredCategory, realizedCategory, vorVehicles)
		}

		// Generate standard ICAO NOTAM text proposal
		now := time.Now().UTC()
		draftNOTAM = fmt.Sprintf("A) %s B) %s C) UFN E) ARFF CAT DOWNGRADED FROM CAT %d TO CAT %d DUE TO FLEET UNSERVICEABILITY. DISCHARGE CAP: %d L/MIN.",
			airportName, now.Format("0601021504"), declaredCategory, realizedCategory, availDischarge)
	}

	return &model.LevelOfProtectionStatus{
		AirportID:              airportID,
		AirportName:            airportName,
		DeclaredCategory:       declaredCategory,
		RealizedCategory:       realizedCategory,
		IsDegraded:             isDegraded,
		ComplianceStatus:       complianceStatus,
		TotalVehicles:          totalVehicles,
		ServiceableVehicles:    serviceableVehicles,
		VORVehicles:            vorVehicles,
		RequiredWaterLiters:    specDeclared.MinWaterL,
		AvailableWaterLiters:   availWater,
		RequiredDischargeRate:  specDeclared.DischargeRate,
		AvailableDischargeRate: availDischarge,
		RequiredDCPKg:          specDeclared.MinDCPKg,
		AvailableDCPKg:         availDCP,
		RequiredVehiclesCount:  specDeclared.MinVehicles,
		AvailableMajorVehicles: availMajorVehicles,
		AlertMessage:           alertMsg,
		DraftNOTAM:             draftNOTAM,
	}
}
