package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type API struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
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

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
	}
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

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode error response")
	}
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
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode paginated response")
	}
}
