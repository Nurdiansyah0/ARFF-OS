package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Nurdiansyah0/arff-os-backend/internal/approval"
	"github.com/Nurdiansyah0/arff-os-backend/internal/middleware"
	"github.com/Nurdiansyah0/arff-os-backend/internal/model"
	"github.com/Nurdiansyah0/arff-os-backend/internal/repository"
)

type ApprovalHandler struct {
	store *repository.Store
}

func NewApprovalHandler(store *repository.Store) *ApprovalHandler {
	return &ApprovalHandler{store: store}
}

// SubmitApproval handles initial submission of an operational item for multi-tier sign-off
func (h *ApprovalHandler) SubmitApproval(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req model.SubmitApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.EntityType == "" || req.EntityID == uuid.Nil || req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "entity_type, entity_id, and title are required"})
		return
	}

	user, err := h.store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Submitting user not found"})
		return
	}

	now := time.Now().UTC()
	appID := uuid.New()
	docHash := approval.HashPayload(req.PayloadJSON)
	if docHash == "" {
		docHash = approval.HashPayload(req.Title + ":" + req.EntityType)
	}

	app := &model.Approval{
		ID:                  appID,
		AirportID:           claims.AirportID,
		EntityType:          req.EntityType,
		EntityID:            req.EntityID,
		Title:               strings.TrimSpace(req.Title),
		Description:         strings.TrimSpace(req.Description),
		CurrentStage:        model.StageSubmitted,
		Status:              model.ApprovalStatusPending,
		SubmittedBy:         user.ID,
		SubmitterName:       user.Name,
		CurrentAssigneeRole: model.RoleSquadLeader,
		DocumentPayloadHash: docHash,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	initialToken := approval.GenerateSignatureToken(appID, user.ID, user.Role, model.StageSubmitted, now)
	initialAction := &model.ApprovalAction{
		ID:             uuid.New(),
		ApprovalID:     appID,
		ActorID:        user.ID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         model.ActionSubmit,
		StageFrom:      "draft",
		StageTo:        model.StageSubmitted,
		Notes:          "Initial submission for review",
		SignatureToken: initialToken,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		CreatedAt:      now,
	}

	if err := h.store.CreateApproval(r.Context(), app, initialAction); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to submit approval request"})
		return
	}

	app.Actions = []*model.ApprovalAction{initialAction}
	writeJSON(w, http.StatusCreated, app)
}

// GetApprovals lists approvals filtered by status
func (h *ApprovalHandler) GetApprovals(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	statusFilter := r.URL.Query().Get("status")
	list, err := h.store.GetApprovalsByAirport(r.Context(), claims.AirportID, statusFilter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approvals"})
		return
	}
	if list == nil {
		list = []*model.Approval{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":     len(list),
		"approvals": list,
	})
}

// GetApprovalDetail returns a single approval and its entire cryptographic audit trail
func (h *ApprovalHandler) GetApprovalDetail(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	idStr := r.PathValue("id")
	appID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid approval ID format"})
		return
	}

	app, err := h.store.GetApprovalByID(r.Context(), claims.AirportID, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Approval request not found"})
		return
	}

	writeJSON(w, http.StatusOK, app)
}

// ProcessAction performs a step in the multi-tier approval matrix
func (h *ApprovalHandler) ProcessAction(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	idStr := r.PathValue("id")
	appID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid approval ID format"})
		return
	}

	var req model.ReviewApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	app, err := h.store.GetApprovalByID(r.Context(), claims.AirportID, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Approval request not found"})
		return
	}

	actor, err := h.store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Actor user not found"})
		return
	}

	actionRecord, err := approval.ValidateAndApplyAction(app, actor, req.Action, req.Notes, r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	if err := h.store.SaveApprovalAction(r.Context(), app, actionRecord); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to record approval action"})
		return
	}

	app.Actions = append(app.Actions, actionRecord)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Approval action recorded successfully",
		"approval": app,
		"action":   actionRecord,
	})
}
