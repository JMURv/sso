package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Password    string       `json:"password"`
	Email       string       `json:"email"`
	Avatar      string       `json:"avatar"`
	Address     string       `json:"address"`
	Phone       string       `json:"phone"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type PaginatedUser struct {
	Data        []*User `json:"data"`
	Count       int64   `json:"count"`
	TotalPages  int     `json:"total_pages"`
	CurrentPage int     `json:"current_page"`
	HasNextPage bool    `json:"has_next_page"`
}

type Client struct {
	ID           uint64   `json:"id"`
	ClientID     string   `json:"client_id"` // UNI
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
}
