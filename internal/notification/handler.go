package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	notifications, total, err := h.svc.List(r.Context(), limit, page)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "failed to list notifications")
		return
	}

	response.Paginated(w, notifications, page, limit, int64(total))
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "notification id is required")
		return
	}

	if err := h.svc.MarkRead(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "notification not found")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "stream_error", "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, err := h.svc.Subscribe(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "stream_error", "failed to subscribe")
		return
	}

	for event := range ch {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}
