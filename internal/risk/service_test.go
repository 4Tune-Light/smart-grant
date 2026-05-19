package risk

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func authCtx(userID, role string) context.Context {
	ctx := context.WithValue(context.Background(), middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func TestScore_Success(t *testing.T) {
	repo := &mockRepository{
		saveFn: func(ctx context.Context, score *RiskScore) error { return nil },
	}
	svc := NewService(repo, &mockProposalRepo{})

	ctx := authCtx("reviewer-1", "reviewer")
	resp, err := svc.Score(ctx, "proposal-1")

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.RiskLevel)
	assert.NotEmpty(t, resp.ProposalID)
}

func TestScore_AdminCanScore(t *testing.T) {
	repo := &mockRepository{
		saveFn: func(ctx context.Context, score *RiskScore) error { return nil },
	}
	svc := NewService(repo, &mockProposalRepo{})

	ctx := authCtx("admin-1", "admin")
	resp, err := svc.Score(ctx, "proposal-1")

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.RiskLevel)
}

func TestScore_NoPermission(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockProposalRepo{})

	ctx := authCtx("applicant-1", "applicant")
	_, err := svc.Score(ctx, "proposal-1")
	assert.Error(t, err)
}

func TestGetScore_Existing(t *testing.T) {
	repo := &mockRepository{
		findByProposalFn: func(ctx context.Context, proposalID string) (*RiskScore, error) {
			return &RiskScore{
				ID: "score-1", ProposalID: proposalID,
				RiskLevel: "medium", Confidence: 0.75,
			}, nil
		},
	}
	svc := NewService(repo, &mockProposalRepo{})

	ctx := authCtx("admin-1", "admin")
	resp, err := svc.GetScore(ctx, "proposal-1")

	assert.NoError(t, err)
	assert.Equal(t, "medium", resp.RiskLevel)
}

func TestGetScore_NotFound(t *testing.T) {
	repo := &mockRepository{
		findByProposalFn: func(ctx context.Context, proposalID string) (*RiskScore, error) {
			return nil, nil
		},
	}
	svc := NewService(repo, &mockProposalRepo{})

	ctx := authCtx("admin-1", "admin")
	_, err := svc.GetScore(ctx, "proposal-1")
	assert.ErrorIs(t, err, ErrScoreNotFound)
}
