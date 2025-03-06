package dto

import (
	"github.com/google/uuid"
	"time"
)

type StartProviderResponse struct {
	URL string `json:"url"`
}

type ProviderResponse struct {
	ID           int       `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Picture      string    `json:"picture"`
	Provider     string    `json:"provider"`
	ProviderID   string    `json:"provider_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	Expiry       time.Time `json:"expiry"`
	ExpiresIn    int64     `json:"expires_in"`
	TokenType    string    `json:"token_type"`
}
