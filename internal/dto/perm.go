package dto

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}
