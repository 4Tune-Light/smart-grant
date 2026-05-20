package dto

import "time"

type DocumentResponse struct {
	ID         string    `json:"id"`
	ProposalID string    `json:"proposal_id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}
