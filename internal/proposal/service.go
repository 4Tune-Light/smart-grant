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
	proposaldto "github.com/rizky/smart-grant/internal/proposal/dto"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/database"
	"github.com/rizky/smart-grant/pkg/storage"
)

type Service interface {
	Create(ctx context.Context, req proposaldto.CreateProposalRequest) (*proposaldto.ProposalResponse, error)
	Update(ctx context.Context, proposalID string, req proposaldto.UpdateProposalRequest) (*proposaldto.ProposalResponse, error)
	Submit(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error)
	GetByID(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error)
	List(ctx context.Context, status string, limit int, cursor *cursor.Cursor) ([]proposaldto.ProposalResponse, *cursor.Cursor, error)
	ListPage(ctx context.Context, status string, limit int, page int) ([]proposaldto.ProposalResponse, int, error)
	UploadDocument(ctx context.Context, proposalID string, file io.Reader, header *multipart.FileHeader) (*proposaldto.DocumentResponse, error)
	GetDocuments(ctx context.Context, proposalID string) ([]proposaldto.DocumentResponse, error)
}

type service struct {
	repo    Repository
	storage storage.FileStorage
	audit   audit.Service
	notif   notification.Service
	tx      *database.Transactor
}

func NewService(repo Repository, st storage.FileStorage, a audit.Service, n notification.Service, tx *database.Transactor) Service {
	return &service{repo: repo, storage: st, audit: a, notif: n, tx: tx}
}

func (s *service) Create(ctx context.Context, req proposaldto.CreateProposalRequest) (*proposaldto.ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	if role != "applicant" {
		return nil, ErrNotApplicant
	}

	var resp *proposaldto.ProposalResponse
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		p := &Proposal{
			ApplicantID:   userID,
			Title:         req.Title,
			Description:   req.Description,
			NominalAmount: req.NominalAmount,
			Organization:  req.Organization,
			Status:        StatusDraft,
			Version:       1,
		}

		if err := s.repo.Create(txCtx, p); err != nil {
			return err
		}

		snapshot, _ := json.Marshal(p)
		s.repo.CreateVersion(txCtx, p.ID, 1, string(snapshot))

		s.audit.Log(txCtx, audit.LogEntry{
			EntityType: "proposal",
			EntityID:   p.ID,
			Action:     "create",
			ActorID:    userID,
			NewValues:  string(snapshot),
		})

		resp = toProposalResp(p)
		return nil
	})

	return resp, err
}

func (s *service) Update(ctx context.Context, proposalID string, req proposaldto.UpdateProposalRequest) (*proposaldto.ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.ApplicantID != userID {
		return nil, ErrNotOwner
	}

	if proposal.Status != StatusDraft {
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

	return toProposalResp(proposal), nil
}

func (s *service) Submit(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error) {
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

	if proposal.Status != StatusDraft {
		return nil, ErrInvalidStatus
	}

	proposal.Status = StatusSubmitted

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

	return toProposalResp(proposal), nil
}

func (s *service) GetByID(ctx context.Context, proposalID string) (*proposaldto.ProposalResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)
	role, _ := ctx.Value(middleware.AuthRoleKey).(string)

	proposal, err := s.repo.FindByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if role == "applicant" && proposal.ApplicantID != userID {
		return nil, ErrNotFound
	}

	return toProposalResp(proposal), nil
}

func (s *service) List(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposaldto.ProposalResponse, *cursor.Cursor, error) {
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

	responses := make([]proposaldto.ProposalResponse, len(proposals))
	for i, p := range proposals {
		responses[i] = *toProposalResp(&p)
	}

	return responses, nextCursor, nil
}

func (s *service) ListPage(ctx context.Context, status string, limit int, page int) ([]proposaldto.ProposalResponse, int, error) {
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

	responses := make([]proposaldto.ProposalResponse, len(proposals))
	for i, p := range proposals {
		responses[i] = *toProposalResp(&p)
	}

	return responses, total, nil
}

func (s *service) UploadDocument(ctx context.Context, proposalID string, file io.Reader, header *multipart.FileHeader) (*proposaldto.DocumentResponse, error) {
	userID, _ := ctx.Value(middleware.AuthUserIDKey).(string)

	if s.storage == nil {
		return nil, fmt.Errorf("file storage not available")
	}

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

	return &proposaldto.DocumentResponse{
		ID:         d.ID,
		ProposalID: d.ProposalID,
		Filename:   d.Filename,
		MimeType:   d.MimeType,
		FileSize:   d.FileSize,
		UploadedAt: d.UploadedAt,
	}, nil
}

func (s *service) GetDocuments(ctx context.Context, proposalID string) ([]proposaldto.DocumentResponse, error) {
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

	responses := make([]proposaldto.DocumentResponse, len(docs))
	for i, d := range docs {
		responses[i] = proposaldto.DocumentResponse{
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

func toProposalResp(p *Proposal) *proposaldto.ProposalResponse {
	return &proposaldto.ProposalResponse{
		ID:            p.ID,
		ApplicantID:   p.ApplicantID,
		Title:         p.Title,
		Description:   p.Description,
		NominalAmount: p.NominalAmount,
		Organization:  p.Organization,
		Status:        string(p.Status),
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
