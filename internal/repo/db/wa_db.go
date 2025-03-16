package db

import (
	"context"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) CreateWACredential(
	ctx context.Context,
	userID uuid.UUID,
	cred *md.WebauthnCredential,
) error {
	const op = "auth.CreateWACredential.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	authenticatorJSON, err := json.Marshal(cred.Authenticator)
	if err != nil {
		zap.L().Error("Failed to marshal authenticator", zap.String("op", op), zap.Error(err))
		return err
	}

	_, err = r.conn.ExecContext(
		ctx, `
		INSERT INTO webauthn_credentials 
		(id, public_key, attestation_type, authenticator, user_id)
		VALUES ($1, $2, $3, $4, $5)`,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		authenticatorJSON,
		userID,
	)
	if err != nil {
		zap.L().Error(
			"Failed to create WebAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *Repository) GetWACredentials(ctx context.Context, userID uuid.UUID) ([]webauthn.Credential, error) {
	const op = "auth.GetWACredentials.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	rows, err := r.conn.QueryContext(
		ctx, `
		SELECT id, public_key, attestation_type, authenticator
		FROM webauthn_credentials
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		zap.L().Error(
			"Failed to query WebAuthn credentials",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	creds := make([]webauthn.Credential, 0, config.DefaultSize)
	for rows.Next() {
		var (
			id              []byte
			publicKey       []byte
			attestationType string
			authenticator   webauthn.Authenticator
			authJSON        []byte
		)

		if err := rows.Scan(&id, &publicKey, &attestationType, &authJSON); err != nil {
			zap.L().Error(
				"Failed to scan WebAuthn credential",
				zap.String("op", op),
				zap.Error(err),
			)
			return nil, err
		}

		if err := json.Unmarshal(authJSON, &authenticator); err != nil {
			zap.L().Error(
				"Failed to unmarshal authenticator",
				zap.String("op", op),
				zap.Error(err),
			)
			return nil, err
		}

		creds = append(
			creds, webauthn.Credential{
				ID:              id,
				PublicKey:       publicKey,
				AttestationType: attestationType,
				Authenticator:   authenticator,
			},
		)
	}

	if err := rows.Err(); err != nil {
		zap.L().Error(
			"Row iteration error",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}

	return creds, nil
}
