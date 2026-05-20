package dto

type RiskResponse struct {
	ID           string             `json:"id"`
	ProposalID   string             `json:"proposal_id"`
	RiskLevel    string             `json:"risk_level"`
	Confidence   float64            `json:"confidence"`
	Features     map[string]float64 `json:"features"`
	ModelVersion string             `json:"model_version"`
	CreatedAt    string             `json:"created_at"`
}
