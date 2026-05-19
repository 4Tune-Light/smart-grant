package proposal

import (
	"context"
	"io"
	"time"

	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/pkg/cursor"
)

type mockRepository struct {
	createFn        func(ctx context.Context, p *Proposal) error
	updateFn        func(ctx context.Context, p *Proposal) error
	findByIDFn      func(ctx context.Context, id string) (*Proposal, error)
	listByAppFn     func(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error)
	listAllFn       func(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error)
	listByAppPageFn func(ctx context.Context, applicantID string, status string, limit int, page int) ([]Proposal, int, error)
	listAllPageFn   func(ctx context.Context, status string, limit int, page int) ([]Proposal, int, error)
}

func (m *mockRepository) Create(ctx context.Context, p *Proposal) error { return m.createFn(ctx, p) }
func (m *mockRepository) Update(ctx context.Context, p *Proposal) error { return m.updateFn(ctx, p) }
func (m *mockRepository) FindByID(ctx context.Context, id string) (*Proposal, error) { return m.findByIDFn(ctx, id) }
func (m *mockRepository) ListByApplicant(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) { return m.listByAppFn(ctx, applicantID, status, limit, c) }
func (m *mockRepository) ListAll(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) { return m.listAllFn(ctx, status, limit, c) }
func (m *mockRepository) ListByApplicantPage(ctx context.Context, applicantID string, status string, limit int, page int) ([]Proposal, int, error) { return m.listByAppPageFn(ctx, applicantID, status, limit, page) }
func (m *mockRepository) ListAllPage(ctx context.Context, status string, limit int, page int) ([]Proposal, int, error) { return m.listAllPageFn(ctx, status, limit, page) }
func (m *mockRepository) CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error { return nil }
func (m *mockRepository) CreateDocument(ctx context.Context, d *Document) error { return nil }
func (m *mockRepository) FindDocuments(ctx context.Context, proposalID string) ([]Document, error) { return nil, nil }

type mockStorage struct{}

func (m *mockStorage) Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error) {
	return "http://minio/test/" + objectPath, nil
}

type mockAudit struct{}

func (m *mockAudit) Log(ctx context.Context, entry audit.LogEntry) error { return nil }
func (m *mockAudit) List(ctx context.Context, filter audit.AuditFilter) ([]audit.AuditResponse, *cursor.Cursor, error) { return nil, nil, nil }

type mockNotif struct{}

func (m *mockNotif) Send(ctx context.Context, userID string, notifType string, title string, body string) error { return nil }
func (m *mockNotif) List(ctx context.Context, limit int, c *cursor.Cursor) ([]notification.NotificationResponse, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockNotif) MarkRead(ctx context.Context, notificationID string) error { return nil }
func (m *mockNotif) Subscribe(ctx context.Context) (<-chan notification.NotificationEvent, error) { return nil, nil }

func testProposal(id, applicantID string) *Proposal {
	return &Proposal{
		ID:            id,
		ApplicantID:   applicantID,
		Title:         "Test Proposal",
		Description:   "A test proposal for testing",
		NominalAmount: 500000000,
		Organization:  "Test Org",
		Status:        "draft",
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
