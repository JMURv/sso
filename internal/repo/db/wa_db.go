package db

import (
	"context"
	md "github.com/JMURv/sso/internal/models"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

func (r *Repository) CreateWebAuthnCredential(ctx context.Context, userID uuid.UUID, cred *md.WebauthnCredential) error {
	_, err := r.conn.ExecContext(ctx, `
        INSERT INTO webauthn_credentials 
        (id, public_key, attestation_type, authenticator, user_id)
        VALUES ($1, $2, $3, $4, $5)`,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		json.Marshal(cred.Authenticator),
		userID,
	)
	return err
}

func (r *Repository) GetWebAuthnCredentials(ctx context.Context, userID uuid.UUID) []md.WebauthnCredential {
	// Query and return credentials
}
