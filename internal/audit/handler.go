package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rizky/smart-grant/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := AuditFilter{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   chi.URLParam(r, "entity_id"),
		ActorID:    r.URL.Query().Get("actor_id"),
		Action:     r.URL.Query().Get("action"),
		Page:       page,
		Limit:      limit,
	}

	if eid := chi.URLParam(r, "entity_id"); eid != "" {
		filter.EntityID = eid
	}

	entries, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "failed to fetch audit logs")
		return
	}

	response.Paginated(w, entries, page, limit, int64(total))
}
