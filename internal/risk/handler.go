package risk

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rizky/smart-grant/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Score(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.Score(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetScore(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.GetScore(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProposalNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "proposal not found")
	case errors.Is(err, ErrScoreNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "risk score not found. trigger scoring first")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}
