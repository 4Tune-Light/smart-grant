package risk

import "time"

type RiskScore struct {
	ID           string
	ProposalID   string
	RiskLevel    string
	Confidence   float64
	Features     string
	Details      string
	ModelVersion string
	CreatedAt    time.Time
}
