package ctrl

import (
	"context"
	"github.com/JMURv/sso/internal/auth"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type AppRepo interface {
	CreateToken(
		ctx context.Context,
		userID uuid.UUID,
		hashedT string,
		expiresAt time.Time,
		device *md.Device,
	) error
	IsTokenValid(ctx context.Context, userID uuid.UUID, d *md.Device, token string) (bool, error)
	RevokeAllTokens(ctx context.Context, userID uuid.UUID) error

	GetUserByOAuth2(ctx context.Context, provider, providerID string) (*md.User, error)
	CreateOAuth2Connection(ctx context.Context, userID uuid.UUID, provider string, data *dto.ProviderResponse) error

	CreateWACredential(
		ctx context.Context,
		userID uuid.UUID,
		cred *md.WebauthnCredential,
	) error
	GetWACredentials(
		ctx context.Context,
		userID uuid.UUID,
	) ([]webauthn.Credential, error)

	GetByDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*md.RefreshToken, error)
	RevokeByDevice(ctx context.Context, userID uuid.UUID, deviceID string) error
	GetUserDevices(ctx context.Context, userID uuid.UUID) ([]md.Device, error)
	DeleteDevice(ctx context.Context, userID uuid.UUID, deviceID string) error

	userRepo
	permRepo
}

type AppCtrl interface {
	GenPair(ctx context.Context, d *dto.DeviceRequest, uid uuid.UUID, p []md.Permission) (dto.TokenPair, error)
	Authenticate(ctx context.Context, d *dto.DeviceRequest, req *dto.EmailAndPasswordRequest) (*dto.TokenPair, error)
	Refresh(ctx context.Context, d *dto.DeviceRequest, req *dto.RefreshRequest) (*dto.TokenPair, error)
	ParseClaims(ctx context.Context, token string) (auth.Claims, error)
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

	GetUserForWA(ctx context.Context, uid uuid.UUID) (*md.WebauthnUser, error)
	StoreWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID, req *webauthn.SessionData) error
	GetWASession(ctx context.Context, sessionType wa.SessionType, userID uuid.UUID) (*webauthn.SessionData, error)
	StoreWACredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error
	GetUserByEmailForWA(ctx context.Context, email string) (*md.WebauthnUser, error)

	IsUserExist(ctx context.Context, email string) (*dto.ExistsUserResponse, error)
	SearchUser(ctx context.Context, query string, page, size int) (*dto.PaginatedUserResponse, error)
	ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUserResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error)
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, u *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *dto.CreatePermissionRequest) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *dto.UpdatePermissionRequest) error
	DeletePerm(ctx context.Context, id uint64) error
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
	SendForgotPasswordEmail(ctx context.Context, token, uid64, toEmail string) error
	SendUserCredentials(_ context.Context, email, pass string) error
}

type Controller struct {
	repo  AppRepo
	au    *auth.Auth
	cache CacheService
	smtp  EmailService
}

func New(repo AppRepo, au *auth.Auth, cache CacheService, smtp EmailService) *Controller {
	return &Controller{
		repo:  repo,
		au:    au,
		cache: cache,
		smtp:  smtp,
	}
}
