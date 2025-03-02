package models

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	ID         uint64    `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"`
	DeviceID   string    `json:"device_id"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Device struct {
	ID         string    `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	DeviceType string    `json:"device_type"`
	OS         string    `json:"os"`
	Browser    string    `json:"browser"`
	UA         string    `json:"ua"`
	IP         string    `json:"ip"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
}
