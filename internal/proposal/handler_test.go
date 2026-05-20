package proposal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rizky/smart-grant/internal/middleware"
	proposaldto "github.com/rizky/smart-grant/internal/proposal/dto"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/stretchr/testify/assert"
)

type mockServiceHandler struct {
	createFn  func(ctx context.Context, req proposaldto.CreateProposalRequest) (*proposaldto.ProposalResponse, error)
	listFn    func(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposaldto.ProposalResponse, *cursor.Cursor, error)
	listPgFn  func(ctx context.Context, status string, limit int, page int) ([]proposaldto.ProposalResponse, int, error)
}

func (m *mockServiceHandler) Create(ctx context.Context, req proposaldto.CreateProposalRequest) (*proposaldto.ProposalResponse, error) { return m.createFn(ctx, req) }
func (m *mockServiceHandler) Update(ctx context.Context, proposalID string, req proposaldto.UpdateProposalRequest) (*proposaldto.ProposalResponse, error) { return nil, nil }
func (m *mockServiceHandler) Submit(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error) { return nil, nil }
func (m *mockServiceHandler) GetByID(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error) { return nil, nil }
func (m *mockServiceHandler) List(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposaldto.ProposalResponse, *cursor.Cursor, error) { return m.listFn(ctx, status, limit, c) }
func (m *mockServiceHandler) ListPage(ctx context.Context, status string, limit int, page int) ([]proposaldto.ProposalResponse, int, error) { return m.listPgFn(ctx, status, limit, page) }
func (m *mockServiceHandler) UploadDocument(ctx context.Context, proposalID string, file io.Reader, header *multipart.FileHeader) (*proposaldto.DocumentResponse, error) { return nil, nil }
func (m *mockServiceHandler) GetDocuments(ctx context.Context, proposalID string) ([]proposaldto.DocumentResponse, error) { return nil, nil }

func authMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.AuthUserIDKey, "user-1")
		ctx = context.WithValue(ctx, middleware.AuthRoleKey, "applicant")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func proposalRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(authMw)
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/page", h.ListPage)
	return r
}

func TestProposalCreateHandler_Success(t *testing.T) {
	svc := &mockServiceHandler{
		createFn: func(ctx context.Context, req proposaldto.CreateProposalRequest) (*proposaldto.ProposalResponse, error) {
			return &proposaldto.ProposalResponse{Title: req.Title, Status: string(StatusDraft)}, nil
		},
	}
	h := NewHandler(svc)
	r := proposalRouter(h)

	body, _ := json.Marshal(proposaldto.CreateProposalRequest{
		Title: "Grant", Description: "Need funding for project X",
		NominalAmount: 100000000, Organization: "Org",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestProposalCreateHandler_InvalidBody(t *testing.T) {
	h := NewHandler(&mockServiceHandler{})
	r := proposalRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProposalListHandler_Cursor(t *testing.T) {
	svc := &mockServiceHandler{
		listFn: func(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposaldto.ProposalResponse, *cursor.Cursor, error) {
			return []proposaldto.ProposalResponse{{ID: "p1", Title: "Test"}}, &cursor.Cursor{LastID: "p1"}, nil
		},
	}
	h := NewHandler(svc)
	r := proposalRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp.Meta)
}

func TestProposalListPageHandler_Success(t *testing.T) {
	svc := &mockServiceHandler{
		listPgFn: func(ctx context.Context, status string, limit int, page int) ([]proposaldto.ProposalResponse, int, error) {
			return []proposaldto.ProposalResponse{{ID: "p1", Title: "Test"}}, 1, nil
		},
	}
	h := NewHandler(svc)
	r := proposalRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/page?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
