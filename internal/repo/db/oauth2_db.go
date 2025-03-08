package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
)

// TODO: Fetch perms for users
func (r *Repository) GetUserByOAuth2(ctx context.Context, provider, providerID string) (*md.User, error) {
	const op = "auth.GetUserByOAuth2.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	var res md.User
	err := r.conn.QueryRowContext(
		ctx, `
        SELECT u.id, u.email, u.name, u.avatar
        FROM users u
        JOIN oauth2_connections oc ON u.id = oc.user_id
        WHERE oc.provider = $1 AND oc.provider_id = $2`,
		provider, providerID,
	).Scan(
		&res.ID,
		&res.Email,
		&res.Name,
		&res.Avatar,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	return &res, nil
}

func (r *Repository) CreateOAuth2Connection(ctx context.Context, userID uuid.UUID, provider string, req *dto.ProviderResponse) error {
	const op = "auth.CreateOAuth2Connection.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(
		ctx, `
        INSERT INTO oauth2_connections 
        (user_id, provider, provider_id, access_token, refresh_token, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, provider, req.ProviderID, req.AccessToken, req.RefreshToken, req.Expiry,
	)
	return err
}
