package approval_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/approval"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

func TestApprovalMatrix_FullChainOfCommand(t *testing.T) {
	appID := uuid.New()
	airportID := uuid.New()

	officer := &model.User{
		ID:        uuid.New(),
		AirportID: airportID,
		Name:      "Officer Dani",
		Role:      model.RoleFirefighter,
	}

	squadLeader := &model.User{
		ID:        uuid.New(),
		AirportID: airportID,
		Name:      "Danton Hendra",
		Role:      model.RoleSquadLeader,
	}

	supervisor := &model.User{
		ID:        uuid.New(),
		AirportID: airportID,
		Name:      "Kasie Bambang",
		Role:      model.RoleOperationsSupervisor,
	}

	chief := &model.User{
		ID:        uuid.New(),
		AirportID: airportID,
		Name:      "Chief ARFF Subagyo",
		Role:      model.RoleHeadOfDepartment,
	}

	docPayload := `{"inspection_id":"123","readiness":"serviceable","turret":"pass"}`
	docHash := approval.HashPayload(docPayload)

	app := &model.Approval{
		ID:                  appID,
		AirportID:           airportID,
		EntityType:          "vehicle_inspection",
		EntityID:            uuid.New(),
		Title:               "Daily Shift Inspection Foam Tender 01",
		CurrentStage:        model.StageSubmitted,
		Status:              model.ApprovalStatusPending,
		SubmittedBy:         officer.ID,
		CurrentAssigneeRole: model.RoleSquadLeader,
		DocumentPayloadHash: docHash,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// 1. Firefighter should NOT be allowed to review or approve
	_, err := approval.ValidateAndApplyAction(app, officer, model.ActionReview, "Trying to self-review", "127.0.0.1", "test-agent")
	if err == nil {
		t.Errorf("Expected unauthorized error when Firefighter tries to review")
	}

	// 2. Squad Leader reviews
	act1, err := approval.ValidateAndApplyAction(app, squadLeader, model.ActionReview, "Armada dan uji pompa terverifikasi baik", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Squad leader review failed: %v", err)
	}
	if app.CurrentStage != model.StageSquadLeaderReviewed {
		t.Errorf("Expected stage %s, got %s", model.StageSquadLeaderReviewed, app.CurrentStage)
	}
	if act1.SignatureToken == "" {
		t.Errorf("Expected signature token for audit trail")
	}

	// 3. Supervisor validates
	act2, err := approval.ValidateAndApplyAction(app, supervisor, model.ActionValidate, "Kesiapan dinas shift valid", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Supervisor validate failed: %v", err)
	}
	if app.CurrentStage != model.StageSupervisorValidated {
		t.Errorf("Expected stage %s, got %s", model.StageSupervisorValidated, app.CurrentStage)
	}
	if act2.Action != model.ActionValidate {
		t.Errorf("Expected action validate, got %s", act2.Action)
	}

	// 4. Head of Department (Chief ARFF) gives final legal seal
	act3, err := approval.ValidateAndApplyAction(app, chief, model.ActionApprove, "Disetujui untuk operasional penerbangan", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Chief ARFF approval failed: %v", err)
	}
	if app.CurrentStage != model.StageChiefApproved {
		t.Errorf("Expected stage %s, got %s", model.StageChiefApproved, app.CurrentStage)
	}
	if app.Status != model.ApprovalStatusApproved {
		t.Errorf("Expected status approved, got %s", app.Status)
	}
	if app.FinalSignatureHash == "" {
		t.Errorf("Expected cryptographic HoD final signature hash")
	}
	if app.SignedAt == nil {
		t.Errorf("Expected SignedAt timestamp to be populated")
	}
	if act3.StageTo != model.StageChiefApproved {
		t.Errorf("Expected act3 stage_to %s, got %s", model.StageChiefApproved, act3.StageTo)
	}
}

func TestApprovalMatrix_RejectionRequiresNotes(t *testing.T) {
	appID := uuid.New()
	airportID := uuid.New()

	squadLeader := &model.User{
		ID:        uuid.New(),
		AirportID: airportID,
		Name:      "Danton Hendra",
		Role:      model.RoleSquadLeader,
	}

	app := &model.Approval{
		ID:                  appID,
		AirportID:           airportID,
		EntityType:          "vehicle_inspection",
		CurrentStage:        model.StageSubmitted,
		Status:              model.ApprovalStatusPending,
		CurrentAssigneeRole: model.RoleSquadLeader,
	}

	// Reject without notes should fail
	_, err := approval.ValidateAndApplyAction(app, squadLeader, model.ActionReject, "", "127.0.0.1", "test-agent")
	if err == nil {
		t.Errorf("Expected error when rejecting without notes")
	}

	// Reject with notes should succeed
	act, err := approval.ValidateAndApplyAction(app, squadLeader, model.ActionReject, "Tekanan pompa utama tidak mencapai 10 bar, ulangi flushing!", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Expected successful rejection with notes: %v", err)
	}
	if app.Status != model.ApprovalStatusRejected {
		t.Errorf("Expected status rejected, got %s", app.Status)
	}
	if act.StageTo != model.StageRejected {
		t.Errorf("Expected stage_to rejected, got %s", act.StageTo)
	}
}
