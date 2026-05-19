package review

import (
	"context"
	"fmt"
	"time"

	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/pkg/cursor"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/notification"
)

type mockRepository struct {
	createFn         func(ctx context.Context, r *Review) error
	findByPropRevFn  func(ctx context.Context, proposalID, reviewerID string) (*Review, error)
	findByPropIDFn   func(ctx context.Context, proposalID string) ([]Review, error)
	updateStatusFn   func(ctx context.Context, proposalID, status string) error
}

var errMockNotSet = fmt.Errorf("mock function not set")

func (m *mockRepository) Create(ctx context.Context, r *Review) error {
	if m.createFn != nil { return m.createFn(ctx, r) }
	return errMockNotSet
}
func (m *mockRepository) FindByProposalAndReviewer(ctx context.Context, proposalID, reviewerID string) (*Review, error) {
	if m.findByPropRevFn != nil { return m.findByPropRevFn(ctx, proposalID, reviewerID) }
	return nil, nil
}
func (m *mockRepository) FindByProposalID(ctx context.Context, proposalID string) ([]Review, error) {
	if m.findByPropIDFn != nil { return m.findByPropIDFn(ctx, proposalID) }
	return nil, nil
}
func (m *mockRepository) UpdateProposalStatus(ctx context.Context, proposalID, status string) error {
	if m.updateStatusFn != nil { return m.updateStatusFn(ctx, proposalID, status) }
	return nil
}

type mockProposalRepo struct{}

func (m *mockProposalRepo) Create(ctx context.Context, p *proposal.Proposal) error { return nil }
func (m *mockProposalRepo) Update(ctx context.Context, p *proposal.Proposal) error { return nil }
func (m *mockProposalRepo) FindByID(ctx context.Context, id string) (*proposal.Proposal, error) {
	return &proposal.Proposal{
		ID: id, Title: "Test", Description: "Test",
		NominalAmount: 500000000, Organization: "Org",
		Status: "submitted", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil
}
func (m *mockProposalRepo) ListByApplicant(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]proposal.Proposal, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockProposalRepo) ListAll(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposal.Proposal, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockProposalRepo) ListByApplicantPage(ctx context.Context, applicantID string, status string, limit int, page int) ([]proposal.Proposal, int, error) { return nil, 0, nil }
func (m *mockProposalRepo) ListAllPage(ctx context.Context, status string, limit int, page int) ([]proposal.Proposal, int, error) { return nil, 0, nil }
func (m *mockProposalRepo) CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error { return nil }
func (m *mockProposalRepo) CreateDocument(ctx context.Context, d *proposal.Document) error { return nil }
func (m *mockProposalRepo) FindDocuments(ctx context.Context, proposalID string) ([]proposal.Document, error) { return nil, nil }

type mockAudit struct{}

func (m *mockAudit) Log(ctx context.Context, entry audit.LogEntry) error { return nil }
func (m *mockAudit) List(ctx context.Context, filter audit.AuditFilter) ([]audit.AuditResponse, *cursor.Cursor, error) { return nil, nil, nil }

type mockNotif struct{}

func (m *mockNotif) Send(ctx context.Context, userID string, notifType string, title string, body string) error { return nil }
func (m *mockNotif) List(ctx context.Context, limit int, c *cursor.Cursor) ([]notification.NotificationResponse, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockNotif) MarkRead(ctx context.Context, notificationID string) error { return nil }
func (m *mockNotif) Subscribe(ctx context.Context) (<-chan notification.NotificationEvent, error) { return nil, nil }
