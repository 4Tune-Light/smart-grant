package risk

import (
	"context"
	"io"
	"time"

	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/pkg/cursor"
)

type mockRepository struct {
	saveFn          func(ctx context.Context, score *RiskScore) error
	findByProposalFn func(ctx context.Context, proposalID string) (*RiskScore, error)
}

func (m *mockRepository) Save(ctx context.Context, score *RiskScore) error {
	return m.saveFn(ctx, score)
}

func (m *mockRepository) FindByProposalID(ctx context.Context, proposalID string) (*RiskScore, error) {
	return m.findByProposalFn(ctx, proposalID)
}

func (m *mockRepository) FindAll(ctx context.Context) ([]RiskScore, error) {
	return nil, nil
}

type mockProposalRepo struct{}

func (m *mockProposalRepo) Create(ctx context.Context, p *proposal.Proposal) error { return nil }
func (m *mockProposalRepo) Update(ctx context.Context, p *proposal.Proposal) error { return nil }
func (m *mockProposalRepo) FindByID(ctx context.Context, id string) (*proposal.Proposal, error) {
	return &proposal.Proposal{
		ID: id, Title: "Test", Description: "Test",
		NominalAmount: 500000000, Organization: "Org",
		Status: proposal.StatusSubmitted, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil
}
func (m *mockProposalRepo) ListByApplicant(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]proposal.Proposal, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockProposalRepo) ListAll(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]proposal.Proposal, *cursor.Cursor, error) { return nil, nil, nil }
func (m *mockProposalRepo) ListByApplicantPage(ctx context.Context, applicantID string, status string, limit int, page int) ([]proposal.Proposal, int, error) { return nil, 0, nil }
func (m *mockProposalRepo) ListAllPage(ctx context.Context, status string, limit int, page int) ([]proposal.Proposal, int, error) { return nil, 0, nil }
func (m *mockProposalRepo) CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error { return nil }
func (m *mockProposalRepo) CreateDocument(ctx context.Context, d *proposal.Document) error { return nil }
func (m *mockProposalRepo) FindDocuments(ctx context.Context, proposalID string) ([]proposal.Document, error) { return nil, nil }
func (m *mockProposalRepo) Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error) { return "", nil }
func (m *mockProposalRepo) CountByOrganization(ctx context.Context, organization string, since time.Time) (int, error) { return 0, nil }
func (m *mockProposalRepo) CountDocuments(ctx context.Context, proposalID string) (int, error) { return 0, nil }
