package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	auditdto "github.com/rizky/smart-grant/internal/audit/dto"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	listFn func(ctx context.Context, filter AuditFilter) ([]auditdto.AuditResponse, *cursor.Cursor, error)
}

func (m *mockService) Log(ctx context.Context, entry LogEntry) error { return nil }
func (m *mockService) List(ctx context.Context, filter AuditFilter) ([]auditdto.AuditResponse, *cursor.Cursor, error) {
	return m.listFn(ctx, filter)
}

func TestAuditListHandler_Success(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter AuditFilter) ([]auditdto.AuditResponse, *cursor.Cursor, error) {
			return []auditdto.AuditResponse{
				{ID: "a1", EntityType: "proposal", Action: "create"},
			}, nil, nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Get("/audit-logs", h.List)
	r.Get("/audit-logs/{entity_id}", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/audit-logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestAuditListHandler_Empty(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter AuditFilter) ([]auditdto.AuditResponse, *cursor.Cursor, error) {
			return []auditdto.AuditResponse{}, nil, nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Get("/audit-logs", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/audit-logs?entity_type=proposal", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
