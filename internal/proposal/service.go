package proposal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/storage"
)

type Service interface {
	Create(ctx context.Context, req CreateProposalRequest) (*ProposalResponse, error)
	Update(ctx context.Context, proposalID string, req UpdateProposalRequest) (*ProposalResponse, error)
	Submit(ctx context.Context, proposalID string) (*ProposalResponse, error)
	GetByID(ctx context.Context, proposalID string) (*ProposalResponse, error)
	List(ctx context.Context, status string, limit int, cursor *cursor.Cursor) ([]ProposalResponse, *cursor.Cursor, error)
	ListPage(ctx context.Context, status string, limit int, page int) ([]ProposalResponse, int, error)
	UploadDocument(ctx context.Context, proposalID string, file io.Reader, header *multipart.FileHeader) (*DocumentResponse, error)
	GetDocuments(ctx context.Context, proposalID string) ([]DocumentResponse, error)
}

type service struct {
	repo    Repository
	storage storage.FileStorage
	audit   audit.Service
	notif   notification.Service
}

func NewService(repo Repository, st storage.FileStorage, a audit.Service, n notification.Service) Service {
	return &service{repo: repo, storage: st, audit: a, notif: n}
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

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "proposal",
		EntityID:   p.ID,
		Action:     "create",
		ActorID:    userID,
		NewValues:  string(snapshot),
	})

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

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "update",
		ActorID:    userID,
		NewValues:  string(snapshot),
	})

	return toProposalResponse(proposal), nil
}

func (s *service) Submit(ctx context.Context, proposalID string) (*ProposalResponse, error) {
	ctx, span := otel.Tracer("smart-grant").Start(ctx, "proposal.Submit")
	defer span.End()

	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	span.SetAttributes(
		attribute.String("proposal.id", proposalID),
		attribute.String("actor.id", userID),
	)

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

	s.audit.Log(ctx, audit.LogEntry{
		EntityType: "proposal",
		EntityID:   proposalID,
		Action:     "submit",
		ActorID:    userID,
		NewValues:  `{"status":"submitted"}`,
	})

	s.notif.Send(ctx, userID, "proposal_submitted",
		"Proposal Submitted",
		fmt.Sprintf("Your proposal '%s' has been submitted for review.", proposal.Title),
	)

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

func (s *service) List(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]ProposalResponse, *cursor.Cursor, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	var proposals []Proposal
	var nextCursor *cursor.Cursor
	var err error

	if role == "applicant" {
		proposals, nextCursor, err = s.repo.ListByApplicant(ctx, userID, status, limit, c)
	} else {
		proposals, nextCursor, err = s.repo.ListAll(ctx, status, limit, c)
	}
	if err != nil {
		return nil, nil, err
	}

	responses := make([]ProposalResponse, len(proposals))
	for i, p := range proposals {
		responses[i] = *toProposalResponse(&p)
	}

	return responses, nextCursor, nil
}

func (s *service) ListPage(ctx context.Context, status string, limit int, page int) ([]ProposalResponse, int, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	var proposals []Proposal
	var total int
	var err error

	if role == "applicant" {
		proposals, total, err = s.repo.ListByApplicantPage(ctx, userID, status, limit, page)
	} else {
		proposals, total, err = s.repo.ListAllPage(ctx, status, limit, page)
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

func (s *service) UploadDocument(ctx context.Context, proposalID string, file io.Reader, header *multipart.FileHeader) (*DocumentResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.ApplicantID != userID {
		return nil, ErrNotOwner
	}

	ext := filepath.Ext(header.Filename)
	objectPath := fmt.Sprintf("proposals/%s/%s%s", proposalID, uuid.New().String(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileURL, err := s.storage.Upload(ctx, objectPath, file, header.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	d := &Document{
		ProposalID: proposalID,
		Filename:   header.Filename,
		FileURL:   fileURL,
		MimeType:   contentType,
		FileSize:   header.Size,
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
