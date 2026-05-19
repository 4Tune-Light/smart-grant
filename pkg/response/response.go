package response

import (
	"encoding/json"
	"net/http"
)

type API struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type CursorMeta struct {
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	resp := API{
		Success: status >= 200 && status < 300,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := API{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func Paginated(w http.ResponseWriter, data interface{}, page int, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := API{
		Success: true,
		Data:    data,
		Meta: &PageMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func CursorPaginated(w http.ResponseWriter, data interface{}, nextCursor string, hasMore bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := API{
		Success: true,
		Data:    data,
		Meta: &CursorMeta{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
