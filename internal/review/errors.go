package review

import "errors"

var (
	ErrNotFound           = errors.New("review not found")
	ErrAlreadyReviewed    = errors.New("you have already reviewed this proposal")
	ErrProposalNotReady   = errors.New("proposal must be submitted before review")
	ErrNotReviewer        = errors.New("only reviewers can submit reviews")
	ErrNotAdmin           = errors.New("only admins can approve or reject proposals")
	ErrProposalAlreadyDecided = errors.New("proposal has already been approved or rejected")
)
