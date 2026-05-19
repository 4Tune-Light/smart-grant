package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var c *cursor.Cursor
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		decoded, err := cursor.Decode(cursorStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid_cursor", "invalid cursor")
			return
		}
		cursorObj := decoded
		c = &cursorObj
	}

	filter := AuditFilter{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   chi.URLParam(r, "entity_id"),
		ActorID:    r.URL.Query().Get("actor_id"),
		Action:     r.URL.Query().Get("action"),
		Limit:      limit,
		Cursor:     c,
	}

	if eid := chi.URLParam(r, "entity_id"); eid != "" {
		filter.EntityID = eid
	}

	entries, nextCursor, err := h.svc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "failed to fetch audit logs")
		return
	}

	nextCursorStr := ""
	if nextCursor != nil {
		nextCursorStr = cursor.Encode(*nextCursor)
	}

	response.CursorPaginated(w, entries, nextCursorStr, nextCursor != nil)
}
