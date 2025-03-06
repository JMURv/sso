package ctrl

import (
	"context"
	"fmt"
	md "github.com/JMURv/sso/internal/models"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"time"
)

func (c *Controller) GetUserForWebAuthn(ctx context.Context, uid uuid.UUID) (*md.WebauthnUser, error) {
	user, err := c.repo.GetUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &md.WebauthnUser{
		ID:          user.ID,
		Email:       user.Email,
		Credentials: c.repo.GetWebAuthnCredentials(ctx, user.ID),
	}, nil
}

func (c *Controller) StoreWebAuthnSession(ctx context.Context, sessionType, userID string, data *webauthn.SessionData) error {
	return c.cache.Set(ctx, 5*time.Minute, fmt.Sprintf("webauthn:%s:%s", sessionType, userID), data)
}
