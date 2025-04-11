package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                uuid.UUID          `json:"id" db:"id"`
	Name              string             `json:"name" db:"name"`
	Password          string             `json:"password" db:"password"`
	Email             string             `json:"email" db:"email"`
	Avatar            string             `json:"avatar" db:"avatar"`
	IsWA              bool               `json:"is_wa" db:"is_wa"`
	IsActive          bool               `json:"is_active" db:"is_active"`
	IsEmailVerified   bool               `json:"is_email_verified" db:"is_email_verified"`
	Roles             []Role             `json:"roles"`
	Oauth2Connections []Oauth2Connection `json:"oauth2_connections"`
	Devices           []Device           `json:"devices"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
}
