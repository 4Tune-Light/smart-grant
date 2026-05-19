package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizky/smart-grant/internal/audit"
	"github.com/rizky/smart-grant/internal/auth"
	"github.com/rizky/smart-grant/internal/middleware"
	"github.com/rizky/smart-grant/internal/notification"
	"github.com/rizky/smart-grant/internal/proposal"
	"github.com/rizky/smart-grant/internal/review"
	"github.com/rizky/smart-grant/internal/risk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type e2eDeps struct {
	pool        *pgxpool.Pool
	authSvc     auth.Service
	proposalSvc proposal.Service
	reviewSvc   review.Service
	riskSvc     risk.Service
	auditSvc    audit.Service
	notifSvc    notification.Service
}

func TestE2E_GrantFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:16-alpine",
			Env: map[string]string{
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "testdb",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	runSchema(t, pool)

	deps := newDeps(pool)

	applicantCtx := withUser(ctx, "applicant-1", "applicant")
	reviewerCtx := withUser(ctx, "reviewer-1", "reviewer")
	adminCtx := withUser(ctx, "admin-1", "admin")

	regResp, err := deps.authSvc.Register(ctx, auth.RegisterRequest{
		Email: "applicant@test.com", Password: "password123",
		Name: "Test Applicant", Role: "applicant",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, regResp.AccessToken)

	loginResp, err := deps.authSvc.Login(ctx, auth.LoginRequest{
		Email: "applicant@test.com", Password: "password123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	prop, err := deps.proposalSvc.Create(applicantCtx, proposal.CreateProposalRequest{
		Title: "Research Grant", Description: "Funding for AI research project",
		NominalAmount: 500000000, Organization: "AI Lab",
	})
	require.NoError(t, err)
	assert.Equal(t, "draft", prop.Status)

	prop, err = deps.proposalSvc.Submit(applicantCtx, prop.ID)
	require.NoError(t, err)
	assert.Equal(t, "submitted", prop.Status)

	_, err = deps.authSvc.Register(ctx, auth.RegisterRequest{
		Email: "reviewer@test.com", Password: "password123",
		Name: "Test Reviewer", Role: "reviewer",
	})
	require.NoError(t, err)

	_, err = deps.authSvc.Register(ctx, auth.RegisterRequest{
		Email: "admin@test.com", Password: "password123",
		Name: "Test Admin", Role: "admin",
	})
	require.NoError(t, err)

	reviewResp, err := deps.reviewSvc.Create(reviewerCtx, prop.ID, review.CreateReviewRequest{
		Score: 85, Comment: "Well-structured proposal with clear objectives",
	})
	require.NoError(t, err)
	assert.Equal(t, 85, reviewResp.Score)

	riskResp, err := deps.riskSvc.Score(adminCtx, prop.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, riskResp.RiskLevel)

	approveResp, err := deps.reviewSvc.Approve(adminCtx, prop.ID)
	require.NoError(t, err)
	assert.Equal(t, "approved", approveResp.Status)

	auditEntries, _, err := deps.auditSvc.List(ctx, audit.AuditFilter{
		EntityType: "proposal", EntityID: prop.ID, Limit: 10,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, auditEntries)
	assert.Equal(t, "proposal", auditEntries[0].EntityType)
}

func withUser(ctx context.Context, userID, role string) context.Context {
	ctx = context.WithValue(ctx, middleware.AuthUserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.AuthRoleKey, role)
	return ctx
}

func newDeps(pool *pgxpool.Pool) *e2eDeps {
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, auth.TokenConfig{
		Secret: "test-secret-that-is-at-least-32-bytes-long!!",
		AccessTTL: 15 * time.Minute,
		RefreshTTL: 72 * time.Hour,
	})

	proposalRepo := proposal.NewRepository(pool)
	notifRepo := notification.NewRepository(pool)
	notifSvc := notification.NewService(notifRepo, nil)
	auditRepo := audit.NewRepository(pool)
	auditSvc := audit.NewService(auditRepo)
	proposalSvc := proposal.NewService(proposalRepo, nil, auditSvc, notifSvc)

	reviewRepo := review.NewRepository(pool)
	reviewSvc := review.NewService(reviewRepo, proposalRepo, auditSvc, notifSvc)

	riskRepo := risk.NewRepository(pool)
	riskSvc := risk.NewService(riskRepo, proposalRepo)

	return &e2eDeps{
		pool: pool, authSvc: authSvc, proposalSvc: proposalSvc,
		reviewSvc: reviewSvc, riskSvc: riskSvc,
		auditSvc: auditSvc, notifSvc: notifSvc,
	}
}

func runSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL CHECK (role IN ('admin','reviewer','applicant')),
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS proposals (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		applicant_id UUID NOT NULL REFERENCES users(id),
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		nominal_amount NUMERIC(15,2) NOT NULL CHECK (nominal_amount > 0),
		organization VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'draft'
			CHECK (status IN ('draft','submitted','in_review','approved','rejected','revision_needed')),
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS proposal_documents (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		proposal_id UUID NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
		filename VARCHAR(255) NOT NULL,
		file_url TEXT NOT NULL DEFAULT '',
		mime_type VARCHAR(100) NOT NULL,
		file_size BIGINT NOT NULL DEFAULT 0,
		uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS proposal_versions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		proposal_id UUID NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
		version_number INT NOT NULL,
		snapshot JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE(proposal_id, version_number)
	);
	CREATE TABLE IF NOT EXISTS reviews (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		proposal_id UUID NOT NULL REFERENCES proposals(id),
		reviewer_id UUID NOT NULL REFERENCES users(id),
		score INT NOT NULL CHECK (score >= 0 AND score <= 100),
		comment TEXT,
		status VARCHAR(50) NOT NULL DEFAULT 'pending'
			CHECK (status IN ('pending','approved','rejected')),
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE(proposal_id, reviewer_id)
	);
	CREATE TABLE IF NOT EXISTS risk_scores (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		proposal_id UUID NOT NULL REFERENCES proposals(id),
		risk_level VARCHAR(50) NOT NULL CHECK (risk_level IN ('low','medium','high')),
		confidence REAL NOT NULL DEFAULT 0.0,
		features JSONB,
		details JSONB,
		model_version VARCHAR(50) NOT NULL DEFAULT 'c4.5-v1',
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		entity_type VARCHAR(100) NOT NULL,
		entity_id UUID NOT NULL,
		action VARCHAR(50) NOT NULL,
		actor_id UUID REFERENCES users(id),
		old_values JSONB,
		new_values JSONB,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS notifications (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id),
		type VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		body TEXT NOT NULL,
		is_read BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);`
	_, err := pool.Exec(context.Background(), schema)
	require.NoError(t, err)
}
