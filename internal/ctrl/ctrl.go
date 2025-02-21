package ctrl

import (
	"context"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/google/uuid"
	"io"
	"time"
)

type AppRepo interface {
	userRepo
	permRepo
}

type AppCtrl interface {
	Authenticate(ctx context.Context, req *dto.EmailAndPasswordRequest) (*dto.EmailAndPasswordResponse, error)
	ParseClaims(ctx context.Context, token string) (map[string]any, error)
	GetUserByToken(ctx context.Context, token string) (*md.User, error)
	SendSupportEmail(ctx context.Context, uid uuid.UUID, theme, text string) error
	CheckForgotPasswordEmail(ctx context.Context, password string, uid uuid.UUID, code int) error
	SendForgotPasswordEmail(ctx context.Context, email string) error
	SendLoginCode(ctx context.Context, email, password string) error
	CheckLoginCode(ctx context.Context, email string, code int) (string, string, error)

	IsUserExist(ctx context.Context, email string) (isExist bool, err error)
	SearchUser(ctx context.Context, query string, page, size int) (*md.PaginatedUser, error)
	ListUsers(ctx context.Context, page, size int) (*md.PaginatedUser, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error)
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, u *md.User, fileName string, bytes []byte) (uuid.UUID, string, string, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *md.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *md.Permission) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *md.Permission) error
	DeletePerm(ctx context.Context, id uint64) error
}

type CacheService interface {
	io.Closer
	GetToStruct(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, t time.Duration, key string, val any)
	Delete(ctx context.Context, key string)
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
	cache CacheService
	smtp  EmailService
}

func New(repo AppRepo, cache CacheService, smtp EmailService) *Controller {
	return &Controller{
		repo:  repo,
		cache: cache,
		smtp:  smtp,
	}
}
