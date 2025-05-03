package ctrl

import (
	"context"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/auth/jwt"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type AppRepo interface {
	authRepo
	oauth2Repo
	waRepo
	userRepo
	permRepo
	roleRepo
	deviceRepo
}

type AppCtrl interface {
	GenPair(ctx context.Context, d *dto.DeviceRequest, uid uuid.UUID, p []md.Role) (dto.TokenPair, error)
	Authenticate(ctx context.Context, d *dto.DeviceRequest, req *dto.EmailAndPasswordRequest) (*dto.TokenPair, error)
	Refresh(ctx context.Context, d *dto.DeviceRequest, req *dto.RefreshRequest) (*dto.TokenPair, error)
	ParseClaims(ctx context.Context, token string) (jwt.Claims, error)
	Logout(ctx context.Context, uid uuid.UUID) error

	CheckForgotPasswordEmail(ctx context.Context, req *dto.CheckForgotPasswordEmailRequest) error
	SendForgotPasswordEmail(ctx context.Context, email string) error
	SendLoginCode(ctx context.Context, d *dto.DeviceRequest, email, password string) (dto.TokenPair, error)
	CheckLoginCode(ctx context.Context, d *dto.DeviceRequest, req *dto.CheckLoginCodeRequest) (*dto.TokenPair, error)

	GetOAuth2AuthURL(ctx context.Context, provider string) (*dto.StartProviderResponse, error)
	HandleOAuth2Callback(ctx context.Context, d *dto.DeviceRequest, provider, code, state string) (*dto.HandleCallbackResponse, error)

	GetOIDCAuthURL(ctx context.Context, provider string) (*dto.StartProviderResponse, error)
	HandleOIDCCallback(ctx context.Context, d *dto.DeviceRequest, provider, code, state string) (*dto.HandleCallbackResponse, error)

	StartRegistration(ctx context.Context, uid uuid.UUID) (*protocol.CredentialCreation, error)
	FinishRegistration(ctx context.Context, uid uuid.UUID, r *http.Request) error
	BeginLogin(ctx context.Context, email string) (*protocol.CredentialAssertion, error)
	FinishLogin(ctx context.Context, email string, d dto.DeviceRequest, r *http.Request) (dto.TokenPair, error)

	GetUserForWA(ctx context.Context, uid uuid.UUID, email string) (*md.WebauthnUser, error)
	StoreWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID, req *webauthn.SessionData) error
	GetWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID) (*webauthn.SessionData, error)

	userCtrl
	permCtrl
	roleCtrl

	ListDevices(ctx context.Context, uid uuid.UUID) ([]md.Device, error)
	GetDevice(ctx context.Context, uid uuid.UUID, dID string) (*md.Device, error)
	GetDeviceByID(ctx context.Context, dID string) (*md.Device, error)
	UpdateDevice(ctx context.Context, uid uuid.UUID, dID string, req *dto.UpdateDeviceRequest) error
	DeleteDevice(ctx context.Context, uid uuid.UUID, dID string) error
}

type S3Service interface {
	UploadFile(ctx context.Context, req *s3.UploadFileRequest) (string, error)
}

type CacheService interface {
	io.Closer
	GetInt(ctx context.Context, key string) (int, error)
	GetStr(ctx context.Context, key string) string
	GetToStruct(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, t time.Duration, key string, val any)
	Delete(ctx context.Context, key string)
	InvalidateKeysByPattern(ctx context.Context, pattern string)
}

type EmailService interface {
	SendLoginEmail(_ context.Context, code int, toEmail string)
	SendForgotPasswordEmail(ctx context.Context, token, uid64, toEmail string)
	SendUserCredentials(_ context.Context, email, pass string)
}

type Controller struct {
	repo  AppRepo
	au    auth.Core
	cache CacheService
	s3    S3Service
	smtp  EmailService
}

func New(repo AppRepo, au auth.Core, cache CacheService, s3 S3Service, smtp EmailService) *Controller {
	return &Controller{
		repo:  repo,
		au:    au,
		cache: cache,
		s3:    s3,
		smtp:  smtp,
	}
}
