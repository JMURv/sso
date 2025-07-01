package dto

import (
	"time"

	"github.com/google/uuid"
)

type StartProviderResponse struct {
	URL string `json:"url"`
}

type HandleCallbackResponse struct {
	Access     string `json:"access"`
	Refresh    string `json:"refresh"`
	SuccessURL string `json:"success_url"`
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

type GoogleResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Picture    string `json:"picture"`
	VerifEmail bool   `json:"verified_email"`
}

type GitHubResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}
