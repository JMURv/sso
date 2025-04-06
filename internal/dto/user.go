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

type CreateUserRequest struct {
	Name     string   `json:"name" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required"`
	Avatar   string   `json:"avatar"`
	IsActive bool     `json:"is_active"`
	IsEmail  bool     `json:"is_email_verified"`
	Roles    []uint64 `json:"roles"`
}

type UpdateUserRequest struct {
	Name     string   `json:"name" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password"`
	Avatar   string   `json:"avatar"`
	IsActive bool     `json:"is_active"`
	IsEmail  bool     `json:"is_email_verified"`
	Roles    []uint64 `json:"roles"`
}

type CreateUserResponse struct {
	ID uuid.UUID `json:"id"`
}

type ExistsUserResponse struct {
	Exists bool `json:"exists"`
}
