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

	rows, err := r.conn.QueryContext(ctx, getWACredentials, userID)
	if err != nil {
		zap.L().Error(
			"failed to get WebAuthn credentials",
			zap.String("op", op),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Error(
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
			zap.L().Error(
				"failed to scan WebAuthn credential",
				zap.String("op", op),
				zap.Error(err),
			)
			return nil, err
		}

		authJSON := webauthn.Authenticator{}
		if err = json.Unmarshal(authenticator, &authJSON); err != nil {
			zap.L().Error(
				"failed to unmarshal authenticator",
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
				Authenticator:   authJSON,
			},
		)
	}

	if err = rows.Err(); err != nil {
		zap.L().Error(
			"failed to get WebAuthn credentials",
			zap.String("op", op),
			zap.Error(err),
		)
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
		zap.L().Error("failed to marshal authenticator", zap.String("op", op), zap.Error(err))
		return err
	}

	_, err = r.conn.ExecContext(
		ctx, createWACredential,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		authenticatorJSON,
		userID,
	)
	if err != nil {
		zap.L().Error(
			"failed to create WebAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	_, err = r.conn.ExecContext(ctx, setIsWA, userID)
	if err != nil {
		zap.L().Error(
			"failed to update user is_wa",
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
		zap.L().Error("failed to marshal authenticator", zap.String("op", op), zap.Error(err))
		return err
	}

	res, err := r.conn.ExecContext(
		ctx, updateWACredentials,
		cred.PublicKey,
		cred.AttestationType,
		authenticatorJSON,
		cred.ID,
	)
	if err != nil {
		zap.L().Error(
			"failed to update WebAuthn credential",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find WebAuthn credential",
			zap.String("op", op),
		)
		return repo.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteWACredential(ctx context.Context, id int, uid uuid.UUID) error {
	const op = "auth.UpdateWACredential.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res, err := r.conn.ExecContext(ctx, deleteWACredential, id, uid)
	if err != nil {
		zap.L().Error(
			"failed to delete WebAuthn credential",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		zap.L().Error(
			"failed to get affected rows",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}

	if aff == 0 {
		zap.L().Debug(
			"failed to find WebAuthn credential",
			zap.String("op", op),
			zap.String("uid", uid.String()),
			zap.Int("id", id),
			zap.Error(err),
		)
		return repo.ErrNotFound
	}
	return nil
}
