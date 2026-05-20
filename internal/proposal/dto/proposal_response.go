package dto

import "time"

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
