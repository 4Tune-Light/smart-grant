package dto

type CreateReviewRequest struct {
	Score   int    `json:"score"   validate:"required,min=0,max=100"`
	Comment string `json:"comment" validate:"required,min=5"`
}
