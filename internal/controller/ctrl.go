package ctrl

import (
	"context"
	md "github.com/JMURv/sso/pkg/model"
	"time"
)

type AppRepo interface {
	userRepo
	permRepo
}

type AuthService interface {
	NewToken(u *md.User, d time.Duration) (string, error)
	VerifyToken(tokenStr string) (map[string]any, error)
}

type CacheService interface {
	GetCode(ctx context.Context, key string) (int, error)
	GetToStruct(ctx context.Context, key string, dest any) error

	Set(ctx context.Context, t time.Duration, key string, val any) error
	Delete(ctx context.Context, key string) error
	Close()

	InvalidateKeysByPattern(ctx context.Context, pattern string)
}

type EmailService interface {
	SendLoginEmail(ctx context.Context, code int, toEmail string) error
	SendForgotPasswordEmail(ctx context.Context, token, uid64, toEmail string) error
	SendSupportEmail(ctx context.Context, u *md.User, theme, text string) error
	SendUserCredentials(_ context.Context, email, pass string) error
	SendOptFile(_ context.Context, email string, filename string, bytes []byte) error
}

type Controller struct {
	repo  AppRepo
	auth  AuthService
	cache CacheService
	smtp  EmailService
}

func New(auth AuthService, repo AppRepo, cache CacheService, smtp EmailService) *Controller {
	return &Controller{
		auth:  auth,
		repo:  repo,
		cache: cache,
		smtp:  smtp,
	}
}
