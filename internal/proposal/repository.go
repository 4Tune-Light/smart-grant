package proposal

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Proposal struct {
	ID            string    `json:"id"`
	ApplicantID   string    `json:"applicant_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	NominalAmount float64   `json:"nominal_amount"`
	Organization  string    `json:"organization"`
	Status        string    `json:"status"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Document struct {
	ID         string    `json:"id"`
	ProposalID string    `json:"proposal_id"`
	Filename   string    `json:"filename"`
	FilePath   string    `json:"file_path"`
	MimeType   string    `json:"mime_type"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type ProposalVersion struct {
	ID            string    `json:"id"`
	ProposalID    string    `json:"proposal_id"`
	VersionNumber int       `json:"version_number"`
	Snapshot      string    `json:"snapshot"`
	CreatedAt     time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, p *Proposal) error
	Update(ctx context.Context, p *Proposal) error
	FindByID(ctx context.Context, id string) (*Proposal, error)
	ListByApplicant(ctx context.Context, applicantID string, status string, limit int, offset int) ([]Proposal, int, error)
	ListAll(ctx context.Context, status string, limit int, offset int) ([]Proposal, int, error)
	CreateVersion(ctx context.Context, proposalID string, versionNumber int, snapshot string) error
	CreateDocument(ctx context.Context, d *Document) error
	FindDocuments(ctx context.Context, proposalID string) ([]Document, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, p *Proposal) error {
	query := `
		INSERT INTO proposals (applicant_id, title, description, nominal_amount, organization, status, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
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

	result, err := r.pool.Exec(ctx, query,
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
	err := r.pool.QueryRow(ctx, query, id).Scan(
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

func (r *repository) ListByApplicant(ctx context.Context, applicantID string, status string, limit int, page int) ([]Proposal, int, error) {
	offset := (page - 1) * limit
	countQuery := `SELECT COUNT(*) FROM proposals WHERE applicant_id = $1`
	args := []interface{}{applicantID}
	argIdx := 2

	if status != "" {
		countQuery += ` AND status = $2`
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals WHERE applicant_id = $1`

	queryArgs := []interface{}{applicantID}
	if status != "" {
		query += ` AND status = $2`
		queryArgs = append(queryArgs, status)
	}

	query += ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	return r.scanProposals(ctx, query, queryArgs, total)
}

func (r *repository) ListAll(ctx context.Context, status string, limit int, page int) ([]Proposal, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	countQuery := `SELECT COUNT(*) FROM proposals`
	query := `
		SELECT id, applicant_id, title, description, nominal_amount,
		       organization, status, version, created_at, updated_at
		FROM proposals`

	if status != "" {
		countQuery += ` WHERE status = $1`
		query += ` WHERE status = $1`
		args = append(args, status)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, limit, offset)

	return r.scanProposals(ctx, query, args, total)
}

func (r *repository) scanProposals(ctx context.Context, query string, args []interface{}, total int) ([]Proposal, int, error) {
	rows, err := r.pool.Query(ctx, query, args...)
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
	_, err := r.pool.Exec(ctx, query, proposalID, versionNumber, snapshot)
	return err
}

func (r *repository) CreateDocument(ctx context.Context, d *Document) error {
	query := `
		INSERT INTO proposal_documents (proposal_id, filename, file_path, mime_type, file_size)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uploaded_at`

	return r.pool.QueryRow(ctx, query,
		d.ProposalID, d.Filename, d.FilePath, d.MimeType, d.FileSize,
	).Scan(&d.ID, &d.UploadedAt)
}

func (r *repository) FindDocuments(ctx context.Context, proposalID string) ([]Document, error) {
	query := `
		SELECT id, proposal_id, filename, file_path, mime_type, file_size, uploaded_at
		FROM proposal_documents WHERE proposal_id = $1
		ORDER BY uploaded_at DESC`

	rows, err := r.pool.Query(ctx, query, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.ProposalID, &d.Filename, &d.FilePath,
			&d.MimeType, &d.FileSize, &d.UploadedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
