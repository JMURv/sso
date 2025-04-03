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
	IsWA              string             `json:"is_wa"`
	IsActive          string             `json:"is_active"`
	IsEmailVerified   string             `json:"is_email_verified"`
	Permissions       []Permission       `json:"permissions"`
	Oauth2Connections []Oauth2Connection `json:"oauth2_connections"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}
