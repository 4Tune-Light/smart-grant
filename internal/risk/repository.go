package risk

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rizky/smart-grant/pkg/database"
)

type Repository interface {
	Save(ctx context.Context, score *RiskScore) error
	FindByProposalID(ctx context.Context, proposalID string) (*RiskScore, error)
	FindAll(ctx context.Context) ([]RiskScore, error)
}

type repository struct {
	q *database.Querier
}

func NewRepository(q *database.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) Save(ctx context.Context, score *RiskScore) error {
	query := `
		INSERT INTO risk_scores (proposal_id, risk_level, confidence, features, details, model_version)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.q.QueryRow(ctx, query,
		score.ProposalID, score.RiskLevel, score.Confidence,
		score.Features, score.Details, score.ModelVersion,
	).Scan(&score.ID, &score.CreatedAt)
}

func (r *repository) FindByProposalID(ctx context.Context, proposalID string) (*RiskScore, error) {
	query := `
		SELECT id, proposal_id, risk_level, confidence, features, details, model_version, created_at
		FROM risk_scores WHERE proposal_id = $1`

	score := &RiskScore{}
	err := r.q.QueryRow(ctx, query, proposalID).Scan(
		&score.ID, &score.ProposalID, &score.RiskLevel, &score.Confidence,
		&score.Features, &score.Details, &score.ModelVersion, &score.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return score, nil
}

func (r *repository) FindAll(ctx context.Context) ([]RiskScore, error) {
	query := `SELECT id, proposal_id, risk_level, confidence, features, details, model_version, created_at FROM risk_scores ORDER BY created_at DESC`
	rows, err := r.q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []RiskScore
	for rows.Next() {
		var s RiskScore
		if err := rows.Scan(&s.ID, &s.ProposalID, &s.RiskLevel, &s.Confidence, &s.Features, &s.Details, &s.ModelVersion, &s.CreatedAt); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}
