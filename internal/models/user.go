package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                uuid.UUID          `json:"id"`
	Name              string             `json:"name"`
	Password          string             `json:"password"`
	Email             string             `json:"email"`
	Avatar            string             `json:"avatar"`
	Permissions       []Permission       `json:"permissions"`
	Oauth2Connections []Oauth2Connection `json:"oauth2_connections"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type Oauth2Connection struct {
	ID           uint64    `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Provider     string    `json:"provider"`
	ProviderID   string    `json:"provider_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
