package approval

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
)

var (
	ErrUnauthorizedRole       = errors.New("user role is not authorized for this approval stage")
	ErrInvalidStageTransition = errors.New("invalid approval stage transition")
	ErrAlreadyConcluded       = errors.New("approval request is already concluded (approved or rejected)")
	ErrRejectionNotesRequired = errors.New("rejection notes are mandatory")
)

const signingKey = "ARFF_OS_TAMPER_PROOF_SECRET_KEY_2026"

// HashPayload generates a SHA-256 checksum for the raw operational payload
func HashPayload(payload string) string {
	h := sha256.New()
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSignatureToken generates an HMAC-SHA256 signature for an audit trail step
func GenerateSignatureToken(approvalID uuid.UUID, actorID uuid.UUID, role string, stage string, timestamp time.Time) string {
	message := fmt.Sprintf("%s:%s:%s:%s:%d", approvalID.String(), actorID.String(), role, stage, timestamp.Unix())
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateFinalHoDSeal generates the cryptographic seal when the Head of Department approves
func GenerateFinalHoDSeal(approvalID uuid.UUID, chiefID uuid.UUID, docHash string, signedAt time.Time) string {
	data := fmt.Sprintf("SEAL:CHIEF_ARFF:%s:%s:%s:%d", approvalID.String(), chiefID.String(), docHash, signedAt.Unix())
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateAndApplyAction validates role permissions and computes the next state in the matrix
func ValidateAndApplyAction(
	app *model.Approval,
	actor *model.User,
	action string,
	notes string,
	ip string,
	userAgent string,
) (*model.ApprovalAction, error) {
	if app.Status == model.ApprovalStatusApproved || app.Status == model.ApprovalStatusRejected {
		return nil, ErrAlreadyConcluded
	}

	now := time.Now().UTC()
	stageFrom := app.CurrentStage
	var stageTo string

	// Handle Rejection
	if action == model.ActionReject {
		if notes == "" {
			return nil, ErrRejectionNotesRequired
		}
		// Any tier at current stage or admin can reject
		if !isRoleAuthorizedForStage(actor.Role, app.CurrentStage) && actor.Role != model.RoleAirportAdmin {
			return nil, ErrUnauthorizedRole
		}

		app.Status = model.ApprovalStatusRejected
		app.CurrentStage = model.StageRejected
		app.UpdatedAt = now

		token := GenerateSignatureToken(app.ID, actor.ID, actor.Role, model.StageRejected, now)
		auditAction := &model.ApprovalAction{
			ID:             uuid.New(),
			ApprovalID:     app.ID,
			ActorID:        actor.ID,
			ActorName:      actor.Name,
			ActorRole:      actor.Role,
			Action:         model.ActionReject,
			StageFrom:      stageFrom,
			StageTo:        model.StageRejected,
			Notes:          notes,
			SignatureToken: token,
			IPAddress:      ip,
			UserAgent:      userAgent,
			CreatedAt:      now,
		}
		return auditAction, nil
	}

	// Multi-Tier Sequential Progression:
	// 1. submitted -> (Squad Leader reviews) -> squad_leader_reviewed
	// 2. squad_leader_reviewed -> (Supervisor validates) -> supervisor_validated
	// 3. supervisor_validated -> (Chief ARFF signs) -> chief_approved
	switch app.CurrentStage {
	case model.StageSubmitted:
		if action != model.ActionReview {
			return nil, fmt.Errorf("expected '%s' action at stage %s", model.ActionReview, app.CurrentStage)
		}
		if actor.Role != model.RoleSquadLeader && actor.Role != model.RoleAirportAdmin {
			return nil, ErrUnauthorizedRole
		}
		stageTo = model.StageSquadLeaderReviewed
		app.CurrentStage = stageTo
		app.CurrentAssigneeRole = model.RoleOperationsSupervisor

	case model.StageSquadLeaderReviewed:
		if action != model.ActionValidate {
			return nil, fmt.Errorf("expected '%s' action at stage %s", model.ActionValidate, app.CurrentStage)
		}
		if actor.Role != model.RoleOperationsSupervisor && actor.Role != model.RoleAirportAdmin {
			return nil, ErrUnauthorizedRole
		}
		stageTo = model.StageSupervisorValidated
		app.CurrentStage = stageTo
		app.CurrentAssigneeRole = model.RoleHeadOfDepartment

	case model.StageSupervisorValidated:
		if action != model.ActionApprove {
			return nil, fmt.Errorf("expected '%s' action at stage %s", model.ActionApprove, app.CurrentStage)
		}
		if actor.Role != model.RoleHeadOfDepartment && actor.Role != model.RoleAirportAdmin {
			return nil, ErrUnauthorizedRole
		}
		stageTo = model.StageChiefApproved
		app.CurrentStage = stageTo
		app.Status = model.ApprovalStatusApproved
		app.CurrentAssigneeRole = "completed"

		// Affix cryptographic HoD digital seal
		app.FinalSignatureHash = GenerateFinalHoDSeal(app.ID, actor.ID, app.DocumentPayloadHash, now)
		app.SignedAt = &now

	default:
		return nil, ErrInvalidStageTransition
	}

	app.UpdatedAt = now
	token := GenerateSignatureToken(app.ID, actor.ID, actor.Role, stageTo, now)

	auditAction := &model.ApprovalAction{
		ID:             uuid.New(),
		ApprovalID:     app.ID,
		ActorID:        actor.ID,
		ActorName:      actor.Name,
		ActorRole:      actor.Role,
		Action:         action,
		StageFrom:      stageFrom,
		StageTo:        stageTo,
		Notes:          notes,
		SignatureToken: token,
		IPAddress:      ip,
		UserAgent:      userAgent,
		CreatedAt:      now,
	}

	return auditAction, nil
}

func isRoleAuthorizedForStage(role string, stage string) bool {
	switch stage {
	case model.StageSubmitted:
		return role == model.RoleSquadLeader || role == model.RoleOperationsSupervisor || role == model.RoleHeadOfDepartment
	case model.StageSquadLeaderReviewed:
		return role == model.RoleOperationsSupervisor || role == model.RoleHeadOfDepartment
	case model.StageSupervisorValidated:
		return role == model.RoleHeadOfDepartment
	default:
		return false
	}
}
