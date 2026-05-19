package proposal

import "errors"

var (
	ErrNotFound       = errors.New("proposal not found")
	ErrNotOwner       = errors.New("only the owner can modify this proposal")
	ErrInvalidStatus  = errors.New("proposal is not in a valid state for this action")
	ErrNotApplicant   = errors.New("only applicants can create proposals")
	ErrNotReviewer    = errors.New("only reviewers can perform this action")
	ErrProposalClosed = errors.New("proposal is already closed")
)
