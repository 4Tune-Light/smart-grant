package review

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/internal/middleware"
	reviewdto "github.com/rizky/smart-grant/internal/review/dto"
	"github.com/stretchr/testify/assert"
)

func authCtx(userID, role string) context.Context {
	ctx := context.WithValue(context.Background(), middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, r *Review) error { return nil },
		updateStatusFn: func(ctx context.Context, proposalID, status string) error { return nil },
	}
	svc := NewService(repo, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("reviewer-1", "reviewer")
	resp, err := svc.Create(ctx, "proposal-1", reviewdto.CreateReviewRequest{Score: 85, Comment: "Good proposal, well researched"})

	assert.NoError(t, err)
	assert.Equal(t, 85, resp.Score)
}

func TestCreate_NotReviewer(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("applicant-1", "applicant")
	_, err := svc.Create(ctx, "proposal-1", reviewdto.CreateReviewRequest{Score: 85, Comment: "Good"})

	assert.ErrorIs(t, err, ErrNotReviewer)
}

func TestCreate_DuplicateReview(t *testing.T) {
	repo := &mockRepository{
		findByPropRevFn: func(ctx context.Context, proposalID, reviewerID string) (*Review, error) {
			return &Review{}, nil
		},
	}
	svc := NewService(repo, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("reviewer-1", "reviewer")
	_, err := svc.Create(ctx, "proposal-1", reviewdto.CreateReviewRequest{Score: 85, Comment: "Good"})

	assert.ErrorIs(t, err, ErrAlreadyReviewed)
}

func TestApprove_AsAdmin(t *testing.T) {
	repo := &mockRepository{
		updateStatusFn: func(ctx context.Context, proposalID, status string) error { return nil },
	}
	svc := NewService(repo, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("admin-1", "admin")
	resp, err := svc.Approve(ctx, "proposal-1")

	assert.NoError(t, err)
	assert.Equal(t, "approved", resp.Status)
}

func TestApprove_NotAdmin(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("reviewer-1", "reviewer")
	_, err := svc.Approve(ctx, "proposal-1")

	assert.ErrorIs(t, err, ErrNotAdmin)
}

func TestReject_AsAdmin(t *testing.T) {
	repo := &mockRepository{
		updateStatusFn: func(ctx context.Context, proposalID, status string) error { return nil },
	}
	svc := NewService(repo, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("admin-1", "admin")
	resp, err := svc.Reject(ctx, "proposal-1")

	assert.NoError(t, err)
	assert.Equal(t, "rejected", resp.Status)
}

func TestReject_NotAdmin(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockProposalRepo{}, &mockAudit{}, &mockNotif{}, nil)

	ctx := authCtx("reviewer-1", "reviewer")
	_, err := svc.Reject(ctx, "proposal-1")

	assert.ErrorIs(t, err, ErrNotAdmin)
}
