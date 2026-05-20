package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	notificationdto "github.com/rizky/smart-grant/internal/notification/dto"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/stretchr/testify/assert"
)

type mockHandlerService struct {
	listFn     func(ctx context.Context, limit int, c *cursor.Cursor) ([]notificationdto.NotificationResponse, *cursor.Cursor, error)
	markReadFn func(ctx context.Context, id string) error
}

func (m *mockHandlerService) Send(ctx context.Context, userID string, notifType string, title string, body string) error {
	return nil
}

func (m *mockHandlerService) List(ctx context.Context, limit int, c *cursor.Cursor) ([]notificationdto.NotificationResponse, *cursor.Cursor, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, c)
	}
	return nil, nil, nil
}

func (m *mockHandlerService) MarkRead(ctx context.Context, notificationID string) error {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, notificationID)
	}
	return nil
}

func (m *mockHandlerService) Subscribe(ctx context.Context) (<-chan notificationdto.NotificationEvent, error) {
	return nil, nil
}

func TestNotifListHandler_Success(t *testing.T) {
	svc := &mockHandlerService{
		listFn: func(ctx context.Context, limit int, c *cursor.Cursor) ([]notificationdto.NotificationResponse, *cursor.Cursor, error) {
			return []notificationdto.NotificationResponse{
				{ID: "n1", Title: "Test", Body: "Hello"},
			}, nil, nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Get("/notifications", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notifications", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestNotifMarkReadHandler_Success(t *testing.T) {
	svc := &mockHandlerService{
		markReadFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Patch("/notifications/read", h.MarkRead)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/notifications/read?id=n1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
