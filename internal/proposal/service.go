package proposal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rizky/smart-grant/internal/middleware"
)

type Service interface {
	Create(ctx context.Context, req CreateProposalRequest) (*ProposalResponse, error)
	Update(ctx context.Context, proposalID string, req UpdateProposalRequest) (*ProposalResponse, error)
	Submit(ctx context.Context, proposalID string) (*ProposalResponse, error)
	GetByID(ctx context.Context, proposalID string) (*ProposalResponse, error)
	List(ctx context.Context, status string, limit int, page int) ([]ProposalResponse, int, error)
	UploadDocument(ctx context.Context, proposalID string, req DocumentUploadRequest) (*DocumentResponse, error)
	GetDocuments(ctx context.Context, proposalID string) ([]DocumentResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateProposalRequest) (*ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	if role != "applicant" {
		return nil, ErrNotApplicant
	}

	p := &Proposal{
		ApplicantID:   userID,
		Title:         req.Title,
		Description:   req.Description,
		NominalAmount: req.NominalAmount,
		Organization:  req.Organization,
		Status:        "draft",
		Version:       1,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	snapshot, _ := json.Marshal(p)
	s.repo.CreateVersion(ctx, p.ID, 1, string(snapshot))

	return toProposalResponse(p), nil
}

func (s *service) Update(ctx context.Context, proposalID string, req UpdateProposalRequest) (*ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.ApplicantID != userID {
		return nil, ErrNotOwner
	}

	if proposal.Status != "draft" {
		return nil, ErrInvalidStatus
	}

	if req.Title != "" {
		proposal.Title = req.Title
	}
	if req.Description != "" {
		proposal.Description = req.Description
	}
	if req.NominalAmount > 0 {
		proposal.NominalAmount = req.NominalAmount
	}
	if req.Organization != "" {
		proposal.Organization = req.Organization
	}

	proposal.Version++

	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}

	snapshot, _ := json.Marshal(proposal)
	s.repo.CreateVersion(ctx, proposal.ID, proposal.Version, string(snapshot))

	return toProposalResponse(proposal), nil
}

func (s *service) Submit(ctx context.Context, proposalID string) (*ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.ApplicantID != userID {
		return nil, ErrNotOwner
	}

	if proposal.Status != "draft" {
		return nil, ErrInvalidStatus
	}

	proposal.Status = "submitted"

	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}

	return toProposalResponse(proposal), nil
}

func (s *service) GetByID(ctx context.Context, proposalID string) (*ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if role == "applicant" && proposal.ApplicantID != userID {
		return nil, ErrNotFound
	}

	return toProposalResponse(proposal), nil
}

func (s *service) List(ctx context.Context, status string, limit int, page int) ([]ProposalResponse, int, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	var proposals []Proposal
	var total int
	var err error

	if role == "applicant" {
		proposals, total, err = s.repo.ListByApplicant(ctx, userID, status, limit, page)
	} else {
		proposals, total, err = s.repo.ListAll(ctx, status, limit, page)
	}
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ProposalResponse, len(proposals))
	for i, p := range proposals {
		responses[i] = *toProposalResponse(&p)
	}

	return responses, total, nil
}

func (s *service) UploadDocument(ctx context.Context, proposalID string, req DocumentUploadRequest) (*DocumentResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.ApplicantID != userID {
		return nil, ErrNotOwner
	}

	filePath := fmt.Sprintf("uploads/%s/%s", proposalID, req.Filename)

	d := &Document{
		ProposalID: proposalID,
		Filename:   req.Filename,
		FilePath:   filePath,
		MimeType:   req.MimeType,
		FileSize:   req.FileSize,
	}

	if err := s.repo.CreateDocument(ctx, d); err != nil {
		return nil, err
	}

	return &DocumentResponse{
		ID:         d.ID,
		ProposalID: d.ProposalID,
		Filename:   d.Filename,
		MimeType:   d.MimeType,
		FileSize:   d.FileSize,
		UploadedAt: d.UploadedAt,
	}, nil
}

func (s *service) GetDocuments(ctx context.Context, proposalID string) ([]DocumentResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if role == "applicant" && proposal.ApplicantID != userID {
		return nil, ErrNotFound
	}

	docs, err := s.repo.FindDocuments(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	responses := make([]DocumentResponse, len(docs))
	for i, d := range docs {
		responses[i] = DocumentResponse{
			ID:         d.ID,
			ProposalID: d.ProposalID,
			Filename:   d.Filename,
			MimeType:   d.MimeType,
			FileSize:   d.FileSize,
			UploadedAt: d.UploadedAt,
		}
	}

	return responses, nil
}

func toProposalResponse(p *Proposal) *ProposalResponse {
	return &ProposalResponse{
		ID:            p.ID,
		ApplicantID:   p.ApplicantID,
		Title:         p.Title,
		Description:   p.Description,
		NominalAmount: p.NominalAmount,
		Organization:  p.Organization,
		Status:        p.Status,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
