package dto

import md "github.com/JMURv/sso/internal/models"

type PaginatedPermissionResponse struct {
	Data        []*md.Permission `json:"data"`
	Count       int64            `json:"count"`
	TotalPages  int              `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	HasNextPage bool             `json:"has_next_page"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}
