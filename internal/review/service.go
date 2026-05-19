package review

import (
	"context"
	"fmt"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/internal/proposal"
)

type Service interface {
	Create(ctx context.Context, proposalID string, req CreateReviewRequest) (*ReviewResponse, error)
	GetByProposal(ctx context.Context, proposalID string) ([]ReviewResponse, error)
	Approve(ctx context.Context, proposalID string) (*ReviewResponse, error)
	Reject(ctx context.Context, proposalID string) (*ReviewResponse, error)
}

type service struct {
	repo         Repository
	proposalRepo proposal.Repository
	audit        audit.Service
	notif        notification.Service
}

func NewService(repo Repository, proposalRepo proposal.Repository, a audit.Service, n notification.Service) Service {
	return &service{repo: repo, proposalRepo: proposalRepo, audit: a, notif: n}
}

func (s *service) Create(ctx context.Context, proposalID string, req CreateReviewRequest) (*ReviewResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	if role != "reviewer" {
		return nil, ErrNotReviewer
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status != "submitted" {
		return nil, ErrProposalNotReady
	}

	existing, err := s.repo.FindByProposalAndReviewer(ctx, proposalID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyReviewed
	}

	rev := &Review{
		ProposalID: proposalID,
		ReviewerID: userID,
		Score:      req.Score,
		Comment:    req.Comment,
		Status:     "pending",
	}

	if err := s.repo.Create(ctx, rev); err != nil {
		return nil, err
	}

	status := "in_review"
	if err := s.repo.UpdateProposalStatus(ctx, proposalID, status); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "review",
		EntityID:   rev.ID,
		Action:     "create",
		ActorID:    userID,
		NewValues:  fmt.Sprintf(`{"proposal_id":"%s","score":%d,"status":"pending"}`, proposalID, req.Score),
	})

	s.notif.Send(ctx, prop.ApplicantID, "review_received",
		"Proposal Reviewed",
		fmt.Sprintf("Your proposal '%s' has been reviewed (score: %d/100)", prop.Title, req.Score),
	)

	return toResponse(rev), nil
}

func (s *service) GetByProposal(ctx context.Context, proposalID string) ([]ReviewResponse, error) {
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

	responses := make([]ReviewResponse, len(reviews))
	for i, rev := range reviews {
		responses[i] = *toResponse(&rev)
	}
	return responses, nil
}

func (s *service) Approve(ctx context.Context, proposalID string) (*ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status == "approved" || prop.Status == "rejected" {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, "approved"); err != nil {
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

	return &ReviewResponse{
		ProposalID: proposalID,
		Status:     "approved",
	}, nil
}

func (s *service) Reject(ctx context.Context, proposalID string) (*ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	prop, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if prop.Status == "approved" || prop.Status == "rejected" {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, "rejected"); err != nil {
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

	return &ReviewResponse{
		ProposalID: proposalID,
		Status:     "rejected",
	}, nil
}

func toResponse(rev *Review) *ReviewResponse {
	return &ReviewResponse{
		ID:           rev.ID,
		ProposalID:   rev.ProposalID,
		ReviewerID:   rev.ReviewerID,
		ReviewerName: rev.ReviewerName,
		Score:        rev.Score,
		Comment:      rev.Comment,
		Status:       rev.Status,
		CreatedAt:    rev.CreatedAt,
		UpdatedAt:    rev.UpdatedAt,
	}
}
