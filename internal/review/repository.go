package review

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rizky/smart-grant/pkg/database"
)

type Repository interface {
	Create(ctx context.Context, r *Review) error
	FindByProposalAndReviewer(ctx context.Context, proposalID string, reviewerID string) (*Review, error)
	FindByProposalID(ctx context.Context, proposalID string) ([]Review, error)
	UpdateProposalStatus(ctx context.Context, proposalID string, status string) error
}

type repository struct {
	q *database.Querier
}

func NewRepository(q *database.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, rev *Review) error {
	query := `
		INSERT INTO reviews (proposal_id, reviewer_id, score, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.q.QueryRow(ctx, query,
		rev.ProposalID, rev.ReviewerID, rev.Score, rev.Comment,
	).Scan(&rev.ID, &rev.CreatedAt, &rev.UpdatedAt)
}

func (r *repository) FindByProposalAndReviewer(ctx context.Context, proposalID string, reviewerID string) (*Review, error) {
	query := `
		SELECT id, proposal_id, reviewer_id, '' as reviewer_name,
		       score, comment, status, created_at, updated_at
		FROM reviews WHERE proposal_id = $1 AND reviewer_id = $2`

	rev := &Review{}
	err := r.q.QueryRow(ctx, query, proposalID, reviewerID).Scan(
		&rev.ID, &rev.ProposalID, &rev.ReviewerID, &rev.ReviewerName,
		&rev.Score, &rev.Comment, &rev.Status, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rev, nil
}

func (r *repository) FindByProposalID(ctx context.Context, proposalID string) ([]Review, error) {
	query := `
		SELECT r.id, r.proposal_id, r.reviewer_id, u.name as reviewer_name,
		       r.score, r.comment, r.status, r.created_at, r.updated_at
		FROM reviews r
		JOIN users u ON u.id = r.reviewer_id
		WHERE r.proposal_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.q.Query(ctx, query, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(
			&rev.ID, &rev.ProposalID, &rev.ReviewerID, &rev.ReviewerName,
			&rev.Score, &rev.Comment, &rev.Status, &rev.CreatedAt, &rev.UpdatedAt,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *repository) UpdateProposalStatus(ctx context.Context, proposalID string, status string) error {
	query := `UPDATE proposals SET status = $1, updated_at = now() WHERE id = $2`
	_, err := r.q.Exec(ctx, query, status, proposalID)
	return err
}
