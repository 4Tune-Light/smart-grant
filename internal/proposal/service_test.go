package proposal

import (
	"context"
	"testing"

	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/stretchr/testify/assert"
)

func userCtx(userID, role string) context.Context {
	ctx := context.WithValue(context.Background(), middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, p *Proposal) error { return nil },
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("applicant-1", "applicant")
	resp, err := svc.Create(ctx, CreateProposalRequest{
		Title: "My Grant", Description: "Need funding for research",
		NominalAmount: 100000000, Organization: "My Org",
	})

	assert.NoError(t, err)
	assert.Equal(t, "My Grant", resp.Title)
	assert.Equal(t, "draft", resp.Status)
}

func TestCreate_NotApplicant(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("reviewer-1", "reviewer")
	_, err := svc.Create(ctx, CreateProposalRequest{
		Title: "My Grant", Description: "Need funding",
		NominalAmount: 100000000, Organization: "My Org",
	})

	assert.ErrorIs(t, err, ErrNotApplicant)
}

func TestUpdate_Owner(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			return testProposal(id, "user-1"), nil
		},
		updateFn: func(ctx context.Context, p *Proposal) error { return nil },
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	resp, err := svc.Update(ctx, "p1", UpdateProposalRequest{Title: "Updated Title"})

	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", resp.Title)
}

func TestUpdate_NotOwner(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			return testProposal(id, "owner-user"), nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("other-user", "applicant")
	_, err := svc.Update(ctx, "p1", UpdateProposalRequest{Title: "Hacked"})

	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdate_WrongStatus(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			p := testProposal(id, "user-1")
			p.Status = "submitted"
			return p, nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	_, err := svc.Update(ctx, "p1", UpdateProposalRequest{Title: "Nope"})

	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestSubmit_Success(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			return testProposal(id, "user-1"), nil
		},
		updateFn: func(ctx context.Context, p *Proposal) error { return nil },
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	resp, err := svc.Submit(ctx, "p1")

	assert.NoError(t, err)
	assert.Equal(t, "submitted", resp.Status)
}

func TestSubmit_NotDraft(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			p := testProposal(id, "user-1")
			p.Status = "submitted"
			return p, nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	_, err := svc.Submit(ctx, "p1")

	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestGetByID_ApplicantSeesOwn(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			return testProposal(id, "user-1"), nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	resp, err := svc.GetByID(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetByID_ApplicantCantSeeOther(t *testing.T) {
	repo := &mockRepository{
		findByIDFn: func(ctx context.Context, id string) (*Proposal, error) {
			return testProposal(id, "other-user"), nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	_, err := svc.GetByID(ctx, "p1")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList_ApplicantFiltered(t *testing.T) {
	called := false
	repo := &mockRepository{
		listByAppFn: func(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) {
			called = true
			assert.Equal(t, "user-1", applicantID)
			return nil, nil, nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("user-1", "applicant")
	_, _, err := svc.List(ctx, "", 10, nil)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestList_AdminSeesAll(t *testing.T) {
	called := false
	repo := &mockRepository{
		listAllFn: func(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) {
			called = true
			return nil, nil, nil
		},
	}
	svc := NewService(repo, &mockStorage{}, &mockAudit{}, &mockNotif{})

	ctx := userCtx("admin-1", "admin")
	_, _, err := svc.List(ctx, "", 10, nil)
	assert.NoError(t, err)
	assert.True(t, called)
}
