package dto

import "time"

type ReviewResponse struct {
	ID           string    `json:"id"`
	ProposalID   string    `json:"proposal_id"`
	ReviewerID   string    `json:"reviewer_id"`
	ReviewerName string    `json:"reviewer_name"`
	Score        int       `json:"score"`
	Comment      string    `json:"comment"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
