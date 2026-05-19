package proposal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	var req CreateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.Create(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	var req UpdateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.Update(r.Context(), proposalID, req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.Submit(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.GetByID(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseIntParam(r.URL.Query().Get("limit"), 10)
	page := parseIntParam(r.URL.Query().Get("page"), 1)

	resp, total, err := h.svc.List(r.Context(), status, limit, page)
	if err != nil {
		handleError(w, err)
		return
	}

	response.Paginated(w, resp, page, limit, int64(total))
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "file field is required")
		return
	}
	defer file.Close()

	resp, err := h.svc.UploadDocument(r.Context(), proposalID, file, header)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	proposalID := chi.URLParam(r, "id")

	resp, err := h.svc.GetDocuments(r.Context(), proposalID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "proposal not found")
	case errors.Is(err, ErrNotOwner):
		response.Error(w, http.StatusForbidden, "not_owner", "you do not own this proposal")
	case errors.Is(err, ErrInvalidStatus):
		response.Error(w, http.StatusConflict, "invalid_status", "proposal is not in a valid state")
	case errors.Is(err, ErrNotApplicant):
		response.Error(w, http.StatusForbidden, "not_applicant", "only applicants can create proposals")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}
