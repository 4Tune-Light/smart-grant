package risk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	riskdto "github.com/rizky/smart-grant/internal/risk/dto"
	"github.com/rizky/smart-grant/pkg/response"
	"github.com/stretchr/testify/assert"
)

type mockHandlerService struct {
	scoreFn    func(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error)
	getScoreFn func(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error)
	retrainFn  func(ctx context.Context) (*RetrainResponse, error)
}

func (m *mockHandlerService) Score(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
	if m.scoreFn != nil {
		return m.scoreFn(ctx, proposalID)
	}
	return nil, nil
}

func (m *mockHandlerService) GetScore(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
	if m.getScoreFn != nil {
		return m.getScoreFn(ctx, proposalID)
	}
	return nil, nil
}

func (m *mockHandlerService) Retrain(ctx context.Context) (*RetrainResponse, error) {
	if m.retrainFn != nil {
		return m.retrainFn(ctx)
	}
	return nil, nil
}

func TestRiskScoreHandler_Success(t *testing.T) {
	svc := &mockHandlerService{
		scoreFn: func(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
			return &riskdto.RiskResponse{
				ProposalID: proposalID, RiskLevel: "low", Confidence: 0.85,
			}, nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/risk/{id}", h.Score)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/risk/p1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.API
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestRiskGetScoreHandler_Success(t *testing.T) {
	svc := &mockHandlerService{
		getScoreFn: func(ctx context.Context, proposalID string) (*riskdto.RiskResponse, error) {
			return &riskdto.RiskResponse{
				ProposalID: proposalID, RiskLevel: "medium", Confidence: 0.72,
			}, nil
		},
	}
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Get("/risk/{id}", h.GetScore)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/risk/p1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
