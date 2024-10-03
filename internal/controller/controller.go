package ctrl

import (
	"context"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	"time"
)

type userRepo interface {
	ListUsers(ctx context.Context, page, size int) (*utils.PaginatedData, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	CreateUser(ctx context.Context, req *model.User) (uuid.UUID, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *model.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UserSearch(ctx context.Context, query string, page int, size int) (*utils.PaginatedData, error)
}

type Auth interface {
	NewToken(u *model.User, d time.Duration) (string, error)
	VerifyToken(tokenStr string) (map[string]any, error)
}

type CacheRepo interface {
	GetCode(ctx context.Context, key string) (int, error)
	GetToStruct(ctx context.Context, key string, dest any) error

	Set(ctx context.Context, t time.Duration, key string, val any) error
	Delete(ctx context.Context, key string) error
	Close()

	InvalidateKeysByPattern(ctx context.Context, pattern string) error
}

type SMTPRepo interface {
	SendLoginEmail(ctx context.Context, code int, toEmail string) error
	SendForgotPasswordEmail(ctx context.Context, token, uid64, toEmail string) error
	SendSupportEmail(ctx context.Context, u *model.User, theme, text string) error
	SendUserCredentials(_ context.Context, email, pass string) error
	SendOptFile(_ context.Context, email string, filename string, bytes []byte) error
}

type Controller struct {
	auth  Auth
	repo  userRepo
	cache CacheRepo
	smtp  SMTPRepo
}

func New(auth Auth, repo userRepo, cache CacheRepo, smtp SMTPRepo) *Controller {
	return &Controller{
		auth:  auth,
		repo:  repo,
		cache: cache,
		smtp:  smtp,
	}
}
