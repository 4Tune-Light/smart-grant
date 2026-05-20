package dto

type CreateProposalRequest struct {
	Title         string  `json:"title"          validate:"required,min=3,max=255"`
	Description   string  `json:"description"    validate:"required,min=10"`
	NominalAmount float64 `json:"nominal_amount" validate:"required,gt=0"`
	Organization  string  `json:"organization"   validate:"required,min=2,max=255"`
}
