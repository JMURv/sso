package models

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	ID         uint64    `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash  string    `json:"token_hash" db:"token_hash"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	Revoked    bool      `json:"revoked" db:"revoked"`
	DeviceID   string    `json:"device_id" db:"device_id"`
	LastUsedAt time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Device struct {
	ID         string    `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Name       string    `json:"name" db:"name"`
	DeviceType string    `json:"device_type" db:"device_type"`
	OS         string    `json:"os" db:"os"`
	Browser    string    `json:"browser" db:"browser"`
	UA         string    `json:"ua" db:"user_agent"`
	IP         string    `json:"ip" db:"ip"`
	LastActive time.Time `json:"last_active" db:"last_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Oauth2Connection struct {
	ID           uint64    `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Provider     string    `json:"provider" db:"provider"`
	ProviderID   string    `json:"provider_id" db:"provider_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	IDToken      string    `json:"id_token" db:"id_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
