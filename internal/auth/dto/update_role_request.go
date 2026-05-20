package dto

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin reviewer applicant"`
}
