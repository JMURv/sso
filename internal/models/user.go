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
	IsWA              bool               `json:"is_wa"`
	IsActive          bool               `json:"is_active"`
	IsEmailVerified   bool               `json:"is_email_verified"`
	Roles             []Role             `json:"roles"`
	Oauth2Connections []Oauth2Connection `json:"oauth2_connections"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}
