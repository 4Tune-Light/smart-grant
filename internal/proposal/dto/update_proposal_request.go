package dto

type UpdateProposalRequest struct {
	Title         string  `json:"title"          validate:"omitempty,min=3,max=255"`
	Description   string  `json:"description"    validate:"omitempty,min=10"`
	NominalAmount float64 `json:"nominal_amount" validate:"omitempty,gt=0"`
	Organization  string  `json:"organization"   validate:"omitempty,min=2,max=255"`
}
