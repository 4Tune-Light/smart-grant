package proposal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rizky/smart-grant/pkg/cursor"
	"github.com/rizky/smart-grant/pkg/database"
)

type Repository interface {
	Create(ctx context.Context, p *Proposal) error
	Update(ctx context.Context, p *Proposal) error
	FindByID(ctx context.Context, id string) (*Proposal, error)
	ListByApplicant(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error)
	ListAll(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error)
	ListByApplicantPage(ctx context.Context, applicantID string, status string, limit int, page int) ([]Proposal, int, error)
	ListAllPage(ctx context.Context, status string, limit int, page int) ([]Proposal, int, error)
	CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error
	CreateDocument(ctx context.Context, d *Document) error
	FindDocuments(ctx context.Context, proposalID string) ([]Document, error)
	CountByOrganization(ctx context.Context, organization string, since time.Time) (int, error)
	CountDocuments(ctx context.Context, proposalID string) (int, error)
}

type repository struct {
	q *database.Querier
}

func NewRepository(q *database.Querier) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, p *Proposal) error {
	query := `
		INSERT INTO proposals (applicant_id, title, description, nominal_amount, organization, status, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.q.QueryRow(ctx, query,
		p.ApplicantID, p.Title, p.Description, p.NominalAmount,
		p.Organization, p.Status, p.Version,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *repository) Update(ctx context.Context, p *Proposal) error {
	query := `
		UPDATE proposals
		SET title = $1, description = $2, nominal_amount = $3,
		    organization = $4, status = $5, version = $6, updated_at = now()
		WHERE id = $7 AND applicant_id = $8`

	result, err := r.q.Exec(ctx, query,
		p.Title, p.Description, p.NominalAmount,
		p.Organization, p.Status, p.Version,
		p.ID, p.ApplicantID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Proposal, error) {
	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals WHERE id = $1`

	p := &Proposal{}
	err := r.q.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.ApplicantID, &p.Title, &p.Description, &p.NominalAmount,
		&p.Organization, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *repository) ListByApplicant(ctx context.Context, applicantID string, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) {
	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals WHERE applicant_id = $1`

	args := []interface{}{applicantID}
	argIdx := 2

	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
		argIdx++
	}

	if c != nil {
		query += fmt.Sprintf(` AND (created_at, id) < ($%d::timestamptz, $%d::uuid)`, argIdx, argIdx+1)
		args = append(args, c.LastCreatedAt, c.LastID)
		argIdx += 2
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, argIdx)
	args = append(args, limit+1)

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var proposals []Proposal
	for rows.Next() {
		var p Proposal
		if err := rows.Scan(
			&p.ID, &p.ApplicantID, &p.Title, &p.Description, &p.NominalAmount,
			&p.Organization, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		proposals = append(proposals, p)
	}

	var nextCursor *cursor.Cursor
	hasMore := len(proposals) > limit
	if hasMore {
		proposals = proposals[:limit]
		last := proposals[len(proposals)-1]
		nextCursor = &cursor.Cursor{LastID: last.ID, LastCreatedAt: last.CreatedAt}
	}

	return proposals, nextCursor, nil
}

func (r *repository) ListAll(ctx context.Context, status string, limit int, c *cursor.Cursor) ([]Proposal, *cursor.Cursor, error) {
	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals`

	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
		argIdx++
	}

	if c != nil {
		prefix := " AND"
		if status == "" {
			prefix = " WHERE"
		}
		query += fmt.Sprintf(`%s (created_at, id) < ($%d::timestamptz, $%d::uuid)`, prefix, argIdx, argIdx+1)
		args = append(args, c.LastCreatedAt, c.LastID)
		argIdx += 2
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, argIdx)
	args = append(args, limit+1)

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var proposals []Proposal
	for rows.Next() {
		var p Proposal
		if err := rows.Scan(
			&p.ID, &p.ApplicantID, &p.Title, &p.Description, &p.NominalAmount,
			&p.Organization, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		proposals = append(proposals, p)
	}

	var nextCursor *cursor.Cursor
	hasMore := len(proposals) > limit
	if hasMore {
		proposals = proposals[:limit]
		last := proposals[len(proposals)-1]
		nextCursor = &cursor.Cursor{LastID: last.ID, LastCreatedAt: last.CreatedAt}
	}

	return proposals, nextCursor, nil
}

func (r *repository) ListByApplicantPage(ctx context.Context, applicantID string, status string, limit int, page int) ([]Proposal, int, error) {
	offset := (page - 1) * limit

	countArgs := []interface{}{applicantID}
	countQuery := `SELECT COUNT(*) FROM proposals WHERE applicant_id = $1`
	if status != "" {
		countQuery += ` AND status = $2`
		countArgs = append(countArgs, status)
	}

	var total int
	if err := r.q.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals WHERE applicant_id = $1`
	args := []interface{}{applicantID}

	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC LIMIT ` + strconv.Itoa(limit) + ` OFFSET ` + strconv.Itoa(offset)

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var proposals []Proposal
	for rows.Next() {
		var p Proposal
		if err := rows.Scan(
			&p.ID, &p.ApplicantID, &p.Title, &p.Description, &p.NominalAmount,
			&p.Organization, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		proposals = append(proposals, p)
	}

	return proposals, total, nil
}

func (r *repository) ListAllPage(ctx context.Context, status string, limit int, page int) ([]Proposal, int, error) {
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM proposals`
	if status != "" {
		countQuery += ` WHERE status = $1`
	}

	var total int
	if status != "" {
		if err := r.q.QueryRow(ctx, countQuery, status).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.q.QueryRow(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals`
	args := []interface{}{}

	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC LIMIT ` + strconv.Itoa(limit) + ` OFFSET ` + strconv.Itoa(offset)

	rows, err := r.q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var proposals []Proposal
	for rows.Next() {
		var p Proposal
		if err := rows.Scan(
			&p.ID, &p.ApplicantID, &p.Title, &p.Description, &p.NominalAmount,
			&p.Organization, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		proposals = append(proposals, p)
	}

	return proposals, total, nil
}

func (r *repository) CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error {
	query := `
		INSERT INTO proposal_versions (proposal_id, version_number, snapshot)
		VALUES ($1, $2, $3)`
	_, err := r.q.Exec(ctx, query, proposalID, versionNumber, snapshot)
	return err
}

func (r *repository) CreateDocument(ctx context.Context, d *Document) error {
	query := `
		INSERT INTO proposal_documents (proposal_id, filename, file_url, mime_type, file_size)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uploaded_at`

	return r.q.QueryRow(ctx, query,
		d.ProposalID, d.Filename, d.FileURL, d.MimeType, d.FileSize,
	).Scan(&d.ID, &d.UploadedAt)
}

func (r *repository) FindDocuments(ctx context.Context, proposalID string) ([]Document, error) {
	query := `
		SELECT id, proposal_id, filename, file_url, mime_type, file_size, uploaded_at
		FROM proposal_documents WHERE proposal_id = $1
		ORDER BY uploaded_at DESC`

	rows, err := r.q.Query(ctx, query, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.ProposalID, &d.Filename, &d.FileURL,
			&d.MimeType, &d.FileSize, &d.UploadedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *repository) CountByOrganization(ctx context.Context, organization string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM proposals WHERE organization = $1 AND created_at >= $2`
	var count int
	err := r.q.QueryRow(ctx, query, organization, since).Scan(&count)
	return count, err
}

func (r *repository) CountDocuments(ctx context.Context, proposalID string) (int, error) {
	query := `SELECT COUNT(*) FROM proposal_documents WHERE proposal_id = $1`
	var count int
	err := r.q.QueryRow(ctx, query, proposalID).Scan(&count)
	return count, err
}

