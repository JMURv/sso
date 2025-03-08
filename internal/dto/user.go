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

type permission struct {
	ID    uint64 `json:"id" validate:"required"`
	Value bool   `json:"value" validate:"required"`
}

type CreateUserRequest struct {
	Name        string       `json:"name" validate:"required"`
	Email       string       `json:"email" validate:"required,email"`
	Password    string       `json:"password" validate:"required"`
	Avatar      string       `json:"avatar"`
	Permissions []permission `json:"permissions" validate:"required"`
}

type UpdateUserRequest struct {
	Name        string       `json:"name" validate:"required"`
	Email       string       `json:"email" validate:"required,email"`
	Password    string       `json:"password" validate:"required"`
	Avatar      string       `json:"avatar"`
	Permissions []permission `json:"permissions" validate:"required"`
}

type CreateUserResponse struct {
	ID uuid.UUID `json:"id"`
}

type ExistsUserResponse struct {
	Exists bool `json:"exists"`
}
