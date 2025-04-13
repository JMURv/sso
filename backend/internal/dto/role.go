package dto

import md "github.com/JMURv/sso/internal/models"

type PaginatedRoleResponse struct {
	Data        []*md.Role `json:"data"`
	Count       int64      `json:"count"`
	TotalPages  int        `json:"total_pages"`
	CurrentPage int        `json:"current_page"`
	HasNextPage bool       `json:"has_next_page"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Permissions []uint64 `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Permissions []uint64 `json:"permissions"`
}
