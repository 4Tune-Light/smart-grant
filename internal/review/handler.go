package review

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/rizky/smart-grant/pkg/validator"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.Create(r.Context(), proposalID, req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetByProposal(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.GetByProposal(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.Approve(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.Reject(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "review not found")
	case errors.Is(err, ErrAlreadyReviewed):
		response.Error(w, http.StatusConflict, "already_reviewed", "you have already reviewed this proposal")
	case errors.Is(err, ErrProposalNotReady):
		response.Error(w, http.StatusConflict, "proposal_not_ready", "proposal must be submitted before review")
	case errors.Is(err, ErrNotReviewer):
		response.Error(w, http.StatusForbidden, "not_reviewer", "only reviewers can submit reviews")
	case errors.Is(err, ErrNotAdmin):
		response.Error(w, http.StatusForbidden, "not_admin", "only admins can approve or reject")
	case errors.Is(err, ErrProposalAlreadyDecided):
		response.Error(w, http.StatusConflict, "already_decided", "proposal has already been decided")
	case errors.Is(err, proposal.ErrNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "proposal not found")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}
