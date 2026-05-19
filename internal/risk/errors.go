package risk

import "errors"

var (
	ErrProposalNotFound = errors.New("proposal not found")
	ErrScoreNotFound    = errors.New("risk score not found")
)
