package proposal

import "time"

type CreateProposalRequest struct {
	Title         string  `json:"title"          validate:"required,min=3,max=255"`
	Description   string  `json:"description"    validate:"required,min=10"`
	NominalAmount float64 `json:"nominal_amount" validate:"required,gt=0"`
	Organization  string  `json:"organization"   validate:"required,min=2,max=255"`
}

type UpdateProposalRequest struct {
	Title         string  `json:"title"          validate:"omitempty,min=3,max=255"`
	Description   string  `json:"description"    validate:"omitempty,min=10"`
	NominalAmount float64 `json:"nominal_amount" validate:"omitempty,gt=0"`
	Organization  string  `json:"organization"   validate:"omitempty,min=2,max=255"`
}

type ProposalResponse struct {
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

type DocumentResponse struct {
	ID         string    `json:"id"`
	ProposalID string    `json:"proposal_id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type DocumentUploadRequest struct {
	Filename string `json:"filename"  validate:"required"`
	MimeType string `json:"mime_type" validate:"required"`
	FileSize int64  `json:"file_size" validate:"required,gt=0"`
}
