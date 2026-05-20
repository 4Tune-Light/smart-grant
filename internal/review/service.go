package review

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/pkg/database"
	reviewdto "github.com/rizky/smart-grant/internal/review/dto"
)

type Service interface {
	Create(ctx context.Context, proposalID string, req reviewdto.CreateReviewRequest) (*reviewdto.ReviewResponse, error)
	GetByProposal(ctx context.Context, proposalID string) ([]reviewdto.ReviewResponse, error)
	Approve(ctx context.Context, proposalID string) (*reviewdto.ReviewResponse, error)
	Reject(ctx context.Context, proposalID string) (*reviewdto.ReviewResponse, error)
}

type service struct {
	repo         Repository
	proposalRepo proposal.Repository
	audit        audit.Service
	notif        notification.Service
	tx           *database.Transactor
}

func NewService(repo Repository, proposalRepo proposal.Repository, a audit.Service, n notification.Service, tx *database.Transactor) Service {
	return &service{repo: repo, proposalRepo: proposalRepo, audit: a, notif: n, tx: tx}
}

func (s *service) Create(ctx context.Context, proposalID string, req reviewdto.CreateReviewRequest) (*reviewdto.ReviewResponse, error) {
	ctx, span := otel.Tracer("smart-grant").Start(ctx, "review.Create")
	defer span.End()

	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	span.SetAttributes(
		attribute.String("proposal.id", proposalID),
		attribute.String("actor.id", userID),
	)

	if role != "reviewer" {
		return nil, ErrNotReviewer
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status != proposal.StatusSubmitted {
		return nil, ErrProposalNotReady
	}

	existing, err := s.repo.FindByProposalAndReviewer(ctx, proposalID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyReviewed
	}

	var resp *reviewdto.ReviewResponse
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		rev := &Review{
			ProposalID: proposalID,
			ReviewerID: userID,
			Score:      req.Score,
			Comment:    req.Comment,
			Status:     ReviewPending,
		}

		if err := s.repo.Create(txCtx, rev); err != nil {
			return err
		}

		status := string(proposal.StatusInReview)
		if err := s.repo.UpdateProposalStatus(txCtx, proposalID, status); err != nil {
			return err
		}

		s.audit.Log(txCtx, audit.LogEntry{
			EntityType: "review",
			EntityID:   rev.ID,
			Action:     "create",
			ActorID:    userID,
			NewValues:  fmt.Sprintf(`{"proposal_id":"%s","score":%d,"status":"pending"}`, proposalID, req.Score),
		})

		resp = toResponse(rev)
		return nil
	})

	return resp, err
}

func (s *service) GetByProposal(ctx context.Context, proposalID string) ([]reviewdto.ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if role == "applicant" && prop.ApplicantID != userID {
		return nil, proposal.ErrNotFound
	}

	reviews, err := s.repo.FindByProposalID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	responses := make([]reviewdto.ReviewResponse, len(reviews))
	for i, rev := range reviews {
		responses[i] = *toResponse(&rev)
	}
	return responses, nil
}

func (s *service) Approve(ctx context.Context, proposalID string) (*reviewdto.ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status == proposal.StatusApproved || prop.Status == proposal.StatusRejected {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, string(proposal.StatusApproved)); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "approve",
		ActorID:    userID,
		NewValues:  fmt.Sprintf(`{"proposal_id":"%s","status":"approved"}`, proposalID),
	})

	s.notif.Send(ctx, prop.ApplicantID, "proposal_approved",
		"Proposal Approved",
		fmt.Sprintf("Your proposal '%s' has been approved.", prop.Title),
	)

	return &reviewdto.ReviewResponse{
		ProposalID: proposalID,
		Status:     "approved",
	}, nil
}

func (s *service) Reject(ctx context.Context, proposalID string) (*reviewdto.ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status == proposal.StatusApproved || prop.Status == proposal.StatusRejected {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, string(proposal.StatusRejected)); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "reject",
		ActorID:    userID,
		NewValues:  fmt.Sprintf(`{"proposal_id":"%s","status":"rejected"}`, proposalID),
	})

	s.notif.Send(ctx, prop.ApplicantID, "proposal_rejected",
		"Proposal Rejected",
		fmt.Sprintf("Your proposal '%s' has been rejected.", prop.Title),
	)

	return &reviewdto.ReviewResponse{
		ProposalID: proposalID,
		Status:     "rejected",
	}, nil
}

func toResponse(rev *Review) *reviewdto.ReviewResponse {
	return &reviewdto.ReviewResponse{
		ID:           rev.ID,
		ProposalID:   rev.ProposalID,
		ReviewerID:   rev.ReviewerID,
		ReviewerName: rev.ReviewerName,
		Score:        rev.Score,
		Comment:      rev.Comment,
		Status:       string(rev.Status),
		CreatedAt:    rev.CreatedAt,
		UpdatedAt:    rev.UpdatedAt,
	}
}
