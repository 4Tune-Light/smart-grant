package review

import (
	"errors"
	"time"
)

var (
	ErrInvalidScore = errors.New("score must be between 0 and 100")
)

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
)

type Review struct {
	ID           string
	ProposalID   string
	ReviewerID   string
	ReviewerName string
	Score        int
	Comment      string
	Status       ReviewStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r *Review) IsValidScore() error {
	if r.Score < 0 || r.Score > 100 {
		return ErrInvalidScore
	}
	return nil
}
