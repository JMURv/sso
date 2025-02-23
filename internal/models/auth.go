package models

import (
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	ID         int64           `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	TokenHash  string          `json:"token_hash"`
	ExpiresAt  time.Time       `json:"expires_at"`
	Revoked    bool            `json:"revoked"`
	DeviceID   string          `json:"device_id"`
	DeviceInfo json.RawMessage `json:"device_info"`
	LastUsedAt time.Time       `json:"last_used_at"`
	CreatedAt  time.Time       `json:"created_at"`
}
