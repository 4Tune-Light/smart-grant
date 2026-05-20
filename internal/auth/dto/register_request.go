package dto

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Role     string `json:"role"     validate:"required,oneof=applicant reviewer"`
}
