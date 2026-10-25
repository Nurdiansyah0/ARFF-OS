package model

import (
	"time"

	"github.com/google/uuid"
)

// Operational Roles Hierarchy
const (
	RoleFirefighter          = "firefighter"           // Level 1: Submitter / Inspector
	RoleSquadLeader          = "squad_leader"          // Level 2: Komandan Regu (Danton)
	RoleOperationsSupervisor = "operations_supervisor" // Level 3: Kepala Seksi Operasi (Kasie)
	RoleHeadOfDepartment     = "head_of_department"    // Level 4: Chief ARFF (Final Digital Sign-off)
	RoleAirportAdmin         = "airport_admin"         // System Administrator
)

// Approval Stages
const (
	StageSubmitted             = "submitted"
	StageSquadLeaderReviewed   = "squad_leader_reviewed"
	StageSupervisorValidated   = "supervisor_validated"
	StageChiefApproved         = "chief_approved"
	StageRejected              = "rejected"
)

// Approval Actions
const (
	ActionSubmit   = "submit"
	ActionReview   = "review"
	ActionValidate = "validate"
	ActionApprove  = "approve"
	ActionReject   = "reject"
)

// Approval Status
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

// Approval represents an operational document requiring multi-tier chain-of-command sign-off
type Approval struct {
	ID                  uuid.UUID `json:"id"`
	AirportID           uuid.UUID `json:"airport_id"`
	EntityType          string    `json:"entity_type"` // 'vehicle_inspection', 'lop_degradation', 'incident_report', 'shift_handover'
	EntityID            uuid.UUID `json:"entity_id"`
	Title               string    `json:"title"`
	Description         string    `json:"description,omitempty"`
	CurrentStage        string    `json:"current_stage"`
	Status              string    `json:"status"` // 'pending', 'approved', 'rejected'
	SubmittedBy         uuid.UUID `json:"submitted_by"`
	SubmitterName       string    `json:"submitter_name,omitempty"`
	CurrentAssigneeRole string    `json:"current_assignee_role"` // Who needs to sign next

	DocumentPayloadHash string     `json:"document_payload_hash"`
	FinalSignatureHash  string     `json:"final_signature_hash,omitempty"`
	SignedAt            *time.Time `json:"signed_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Audit Trail Actions
	Actions []*ApprovalAction `json:"actions,omitempty"`
}

// ApprovalAction represents one step in the cryptographic audit trail
type ApprovalAction struct {
	ID             uuid.UUID `json:"id"`
	ApprovalID     uuid.UUID `json:"approval_id"`
	ActorID        uuid.UUID `json:"actor_id"`
	ActorName      string    `json:"actor_name"`
	ActorRole      string    `json:"actor_role"`
	Action         string    `json:"action"` // 'submit', 'review', 'validate', 'approve', 'reject'
	StageFrom      string    `json:"stage_from"`
	StageTo        string    `json:"stage_to"`
	Notes          string    `json:"notes,omitempty"`
	SignatureToken string    `json:"signature_token"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// SubmitApprovalRequest is the payload when an officer submits an operational item
type SubmitApprovalRequest struct {
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PayloadJSON string    `json:"payload_json"` // Raw string representation of the data to hash
}

// ReviewApprovalRequest is the payload when a leader/supervisor/chief acts on an approval
type ReviewApprovalRequest struct {
	Action string `json:"action"` // 'review', 'validate', 'approve', 'reject'
	Notes  string `json:"notes"`
}
