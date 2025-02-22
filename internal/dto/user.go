package dto

import (
	md "github.com/JMURv/sso/internal/models"
	"github.com/google/uuid"
)

type PaginatedUserResponse struct {
	Data        []*md.User `json:"data"`
	Count       int64      `json:"count"`
	TotalPages  int        `json:"total_pages"`
	CurrentPage int        `json:"current_page"`
	HasNextPage bool       `json:"has_next_page"`
}

type CreateUserResponse struct {
	ID uuid.UUID `json:"id"`
}

type ExistsUserResponse struct {
	Exists bool `json:"exists"`
}
