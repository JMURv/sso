package models

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type WebauthnUser struct {
	ID          uuid.UUID
	Email       string
	Permissions []Permission
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

func (u *WebauthnUser) ExcludeCredentialDescriptorList() []protocol.CredentialDescriptor {
	ex := make([]protocol.CredentialDescriptor, 0, len(u.Credentials))
	for i := 0; i < len(u.Credentials); i++ {
		ex = append(
			ex, protocol.CredentialDescriptor{
				Type:         protocol.PublicKeyCredentialType,
				CredentialID: u.Credentials[i].ID,
			},
		)
	}
	return ex
}
