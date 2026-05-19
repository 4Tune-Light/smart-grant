package auth

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	resp, err := h.svc.RefreshToken(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.svc.ListUsers(r.Context(), role, limit, page)
	if err != nil {
		handleError(w, err)
		return
	}

	response.Paginated(w, users, page, limit, int64(total))
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if errs := validator.Struct(req); errs != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", errs[0].Message)
		return
	}

	if err := h.svc.UpdateRole(r.Context(), targetID, req.Role); err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
	case errors.Is(err, ErrEmailAlreadyExists):
		response.Error(w, http.StatusConflict, "email_exists", "email already registered")
	case errors.Is(err, ErrInvalidToken):
		response.Error(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
	case errors.Is(err, ErrUserInactive):
		response.Error(w, http.StatusForbidden, "user_inactive", "account is deactivated")
	case errors.Is(err, ErrUserNotFound):
		response.Error(w, http.StatusNotFound, "not_found", "user not found")
	case err != nil && err.Error() == "cannot change your own role":
		response.Error(w, http.StatusForbidden, "cannot_change_own_role", "you cannot change your own role")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}
