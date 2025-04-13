package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func (r *Repository) GetUserByOAuth2(ctx context.Context, provider, providerID string) (*md.User, error) {
	const op = "auth.GetUserByOAuth2.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	res := &md.User{}
	roles := make([]string, 0, 5)
	err := r.conn.QueryRowContext(ctx, getUserOauth2, provider, providerID).Scan(
		&res.ID, &res.Email, &res.Name, &res.Avatar, pq.Array(&roles),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Debug(
				"no user found",
				zap.String("op", op),
				zap.String("provider", provider),
				zap.String("providerID", providerID),
			)
			return nil, repo.ErrNotFound
		}
		zap.L().Error(
			"failed to get user",
			zap.String("op", op),
			zap.String("provider", provider),
			zap.String("providerID", providerID),
			zap.Error(err),
		)
		return nil, err
	}

	res.Roles, err = ScanRoles(roles)
	if err != nil {
		zap.L().Error(
			"failed to scan roles",
			zap.String("op", op),
			zap.Error(err),
		)
		return nil, err
	}

	return res, nil
}

func (r *Repository) CreateOAuth2Connection(ctx context.Context, userID uuid.UUID, provider string, req *dto.ProviderResponse) error {
	const op = "auth.CreateOAuth2Connection.repo"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	_, err := r.conn.ExecContext(
		ctx, createOAuth2Connection,
		userID, provider, req.ProviderID, req.AccessToken, req.RefreshToken, req.Expiry,
	)
	if err != nil {
		zap.L().Error(
			"failed to create oauth2 connection",
			zap.String("op", op),
			zap.String("userID", userID.String()),
			zap.String("provider", provider),
			zap.Error(err),
		)
	}
	return err
}
