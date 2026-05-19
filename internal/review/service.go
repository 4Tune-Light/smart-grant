package review

import (
	"context"

	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/proposal"
)

type Service interface {
	Create(ctx context.Context, proposalID string, req CreateReviewRequest) (*ReviewResponse, error)
	GetByProposal(ctx context.Context, proposalID string) ([]ReviewResponse, error)
	Approve(ctx context.Context, proposalID string) (*ReviewResponse, error)
	Reject(ctx context.Context, proposalID string) (*ReviewResponse, error)
}

type service struct {
	repo      Repository
	proposalRepo proposal.Repository
}

func NewService(repo Repository, proposalRepo proposal.Repository) Service {
	return &service{repo: repo, proposalRepo: proposalRepo}
}

func (s *service) Create(ctx context.Context, proposalID string, req CreateReviewRequest) (*ReviewResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	if role != "reviewer" {
		return nil, ErrNotReviewer
	}

	proposal, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.Status != "submitted" {
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
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	proposal, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.Status == "approved" || proposal.Status == "rejected" {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, "approved"); err != nil {
		return nil, err
	}

	return &ReviewResponse{
		ProposalID: proposalID,
		Status:     "approved",
	}, nil
}

func (s *service) Reject(ctx context.Context, proposalID string) (*ReviewResponse, error) {
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)
	if role != "admin" {
		return nil, ErrNotAdmin
	}

	proposal, err := s.proposalRepo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.Status == "approved" || proposal.Status == "rejected" {
		return nil, ErrProposalAlreadyDecided
	}

	if err := s.repo.UpdateProposalStatus(ctx, proposalID, "rejected"); err != nil {
		return nil, err
	}

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
