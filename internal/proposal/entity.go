package proposal

import (
	"errors"
	"time"
)

type ProposalStatus string

const (
	StatusDraft     ProposalStatus = "draft"
	StatusSubmitted ProposalStatus = "submitted"
	StatusInReview  ProposalStatus = "in_review"
	StatusApproved  ProposalStatus = "approved"
	StatusRejected  ProposalStatus = "rejected"
)

type Proposal struct {
	ID            string
	ApplicantID   string
	Title         string
	Description   string
	NominalAmount float64
	Organization  string
	Status        ProposalStatus
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (p *Proposal) IsOwner(userID string) bool {
	return p.ApplicantID == userID
}

func (p *Proposal) CanSubmit() error {
	if p.Status != StatusDraft {
		return ErrInvalidStatus
	}
	return nil
}

func (p *Proposal) CanUpdate(userID string) error {
	if !p.IsOwner(userID) {
		return ErrNotOwner
	}
	if p.Status != StatusDraft {
		return ErrInvalidStatus
	}
	return nil
}

func (p *Proposal) CanBeReviewed() error {
	if p.Status != StatusSubmitted {
		return ErrInvalidStatus
	}
	return nil
}

func (p *Proposal) CanBeApproved() error {
	if p.Status == StatusApproved || p.Status == StatusRejected {
		return errors.New("proposal has already been decided")
	}
	return nil
}

func (p *Proposal) IsDecided() bool {
	return p.Status == StatusApproved || p.Status == StatusRejected
}

type Document struct {
	ID         string
	ProposalID string
	Filename   string
	FileURL    string
	MimeType   string
	FileSize   int64
	UploadedAt time.Time
}

type ProposalVersion struct {
	ID            string
	ProposalID    string
	VersionNumber int
	Snapshot      string
	CreatedAt     time.Time
}
