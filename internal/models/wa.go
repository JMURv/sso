package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type WebauthnUser struct {
	ID          uuid.UUID
	Email       string
	Credentials []webauthn.Credential
}

type WebauthnCredential struct {
	ID              []byte
	PublicKey       []byte
	AttestationType string
	Authenticator   webauthn.Authenticator
	UserID          uuid.UUID
}

func (u *WebauthnUser) WebAuthnID() []byte {
	return u.ID[:]
}

func (u *WebauthnUser) WebAuthnName() string {
	return u.Email
}

func (u *WebauthnUser) WebAuthnDisplayName() string {
	return u.Email
}

func (u *WebauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *WebauthnUser) WebAuthnIcon() string {
	return ""
}
