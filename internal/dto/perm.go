package dto

type CreatePermissionRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdatePermissionRequest struct {
	Name string `json:"name" validate:"required"`
}
