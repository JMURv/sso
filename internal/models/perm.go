package models

import "github.com/google/uuid"

type Permission struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type PaginatedPermission struct {
	Data        []*Permission `json:"data"`
	Count       int64         `json:"count"`
	TotalPages  int           `json:"total_pages"`
	CurrentPage int           `json:"current_page"`
	HasNextPage bool          `json:"has_next_page"`
}

type UserPermission struct {
	UserID uuid.UUID `json:"user_id"`
	PermID uint64    `json:"permission_id"`
	Value  bool      `json:"value"`
}
