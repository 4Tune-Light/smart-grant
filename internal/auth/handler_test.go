package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authdto "github.com/rizky/smart-grant/internal/auth/dto"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	registerFn    func(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error)
	loginFn       func(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error)
	refreshFn     func(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error)
}

func (m *mockService) Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error) { return m.registerFn(ctx, req) }
func (m *mockService) Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error) { return m.loginFn(ctx, req) }
func (m *mockService) RefreshToken(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error) { return m.refreshFn(ctx, req) }
func (m *mockService) ListUsers(ctx context.Context, role string, limit int, page int) ([]authdto.UserInfo, int, error) { return nil, 0, nil }
func (m *mockService) UpdateRole(ctx context.Context, targetID, newRole string) error { return nil }

func TestRegisterHandler_Success(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error) {
			return &authdto.AuthResponse{
				AccessToken: "token-abc", TokenType: "Bearer",
				User: authdto.UserInfo{Email: req.Email, Role: req.Role},
			}, nil
		},
	}
	h := NewHandler(svc)

	body, _ := json.Marshal(authdto.RegisterRequest{Email: "a@b.com", Password: "pass1234", Name: "Test User", Role: "applicant"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	h.Register(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	h := NewHandler(&mockService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("{invalid")))
	r.Header.Set("Content-Type", "application/json")

	h.Register(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error) {
			return &authdto.AuthResponse{AccessToken: "token-abc", TokenType: "Bearer", User: authdto.UserInfo{Email: req.Email}}, nil
		},
	}
	h := NewHandler(svc)

	body, _ := json.Marshal(authdto.LoginRequest{Email: "a@b.com", Password: "pass"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	h.Login(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error) {
			return nil, ErrInvalidCredentials
		},
	}
	h := NewHandler(svc)

	body, _ := json.Marshal(authdto.LoginRequest{Email: "a@b.com", Password: "wrong"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	h.Login(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
