package db

import (
	"context"
	"database/sql"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/repo"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) GetWACredentials(ctx context.Context, userID uuid.UUID) ([]webauthn.Credential, error) {
	const op = "auth.GetWACredentials.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	rows, err := r.conn.QueryContext(
		ctx, `
		SELECT 
		    id, 
		    public_key,
		    attestation_type,
		    authenticator
		FROM wa_credentials
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Debug(
				"failed to close rows",
				zap.String("op", op),
				zap.Error(err),
			)
		}
	}(rows)

	creds := make([]webauthn.Credential, 0, config.DefaultSize)
	for rows.Next() {
		var (
			id              []byte
			publicKey       []byte
			attestationType string
			authenticator   []byte
		)

		if err = rows.Scan(
			&id,
			&publicKey,
			&attestationType,
			&authenticator,
		); err != nil {
			return nil, err
		}

		authJSON := webauthn.Authenticator{}
		if err = json.Unmarshal(authenticator, &authJSON); err != nil {
			return nil, err
		}

		creds = append(
			creds, webauthn.Credential{
				ID:              id,
				PublicKey:       publicKey,
				AttestationType: attestationType,
				Authenticator:   authJSON,
			},
		)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *Repository) CreateWACredential(ctx context.Context, userID uuid.UUID, cred *webauthn.Credential) error {
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
		INSERT INTO wa_credentials 
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

	_, err = r.conn.ExecContext(ctx, `UPDATE users SET is_wa = TRUE WHERE id = $1`, userID)
	if err != nil {
		zap.L().Error(
			"Failed to update user is_wa",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *Repository) UpdateWACredential(ctx context.Context, cred *webauthn.Credential) error {
	const op = "auth.UpdateWACredential.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	authenticatorJSON, err := json.Marshal(cred.Authenticator)
	if err != nil {
		zap.L().Error("Failed to marshal authenticator", zap.String("op", op), zap.Error(err))
		return err
	}

	res, err := r.conn.ExecContext(
		ctx, `
		UPDATE wa_credentials
		SET public_key = $1,
			attestation_type = $2,
			authenticator = $3
		WHERE id = $4;`,
		cred.PublicKey,
		cred.AttestationType,
		authenticatorJSON,
		cred.ID,
	)
	if err != nil {
		zap.L().Error(
			"Failed to update WebAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteWACredential(ctx context.Context, id int, uid uuid.UUID) error {
	const op = "auth.UpdateWACredential.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, `DELETE FROM wa_credentials WHERE id = $1 AND user_id = $2`, id, uid)
	if err != nil {
		zap.L().Error(
			"Failed to delete WebAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if aff == 0 {
		return repo.ErrNotFound
	}
	return nil
}
