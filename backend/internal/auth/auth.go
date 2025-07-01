package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/JMURv/sso/internal/auth/captcha"
	"github.com/JMURv/sso/internal/auth/jwt"
	"github.com/JMURv/sso/internal/auth/providers"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Core interface {
	Hash(val string) (string, error)
	ComparePasswords(hashed, pswd []byte) error
	jwt.Port
	captcha.Port
	providers.Port
	wa.Port
}

type Auth struct {
	jwt       jwt.Port
	captcha   captcha.Port
	providers providers.Port
	wa        wa.Port
}

func New(conf config.Config) *Auth {
	return &Auth{
		jwt:       jwt.New(conf),
		captcha:   captcha.New(conf),
		providers: providers.New(conf),
		wa:        wa.New(conf),
	}
}

func (a *Auth) Hash(val string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.MinCost)
	if err != nil {
		zap.L().Error(
			"Failed to generate hash",
			zap.String("val", val),
			zap.Error(err),
		)

		return "", err
	}

	return string(bytes), nil
}

func (a *Auth) ComparePasswords(hashed, pswd []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashed, pswd); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

func (a *Auth) GetAccessTime() time.Time {
	return a.jwt.GetAccessTime()
}

func (a *Auth) GetRefreshTime() time.Time {
	return a.jwt.GetRefreshTime()
}

func (a *Auth) GenPair(ctx context.Context, uid uuid.UUID, roles []md.Role) (string, string, error) {
	return a.jwt.GenPair(ctx, uid, roles)
}

func (a *Auth) NewToken(ctx context.Context, uid uuid.UUID, roles []md.Role, d time.Duration) (string, error) {
	return a.jwt.NewToken(ctx, uid, roles, d)
}

func (a *Auth) ParseClaims(ctx context.Context, tokenStr string) (jwt.Claims, error) {
	return a.jwt.ParseClaims(ctx, tokenStr)
}

func (a *Auth) VerifyRecaptcha(token string, action captcha.Actions) (bool, error) {
	return a.captcha.VerifyRecaptcha(token, action)
}

func (a *Auth) Get(provider providers.Providers, flow providers.Flow) (providers.OAuthProvider, error) {
	return a.providers.Get(provider, flow)
}

func (a *Auth) SuccessURL() string {
	return a.providers.SuccessURL()
}

func (a *Auth) GenerateSignedState() string {
	return a.providers.GenerateSignedState()
}

func (a *Auth) ValidateSignedState(signedState string, maxAge time.Duration) error {
	return a.providers.ValidateSignedState(signedState, maxAge)
}

func (a *Auth) BeginLogin(
	user webauthn.User,
	opts ...webauthn.LoginOption,
) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return a.wa.BeginLogin(user, opts...)
}

func (a *Auth) FinishLogin(
	user webauthn.User,
	session webauthn.SessionData,
	response *http.Request,
) (*webauthn.Credential, error) {
	return a.wa.FinishLogin(user, session, response)
}

func (a *Auth) BeginRegistration(
	user webauthn.User,
	opts ...webauthn.RegistrationOption,
) (creation *protocol.CredentialCreation, session *webauthn.SessionData, err error) {
	return a.wa.BeginRegistration(user, opts...)
}

func (a *Auth) FinishRegistration(
	user webauthn.User,
	session webauthn.SessionData,
	response *http.Request,
) (*webauthn.Credential, error) {
	return a.wa.FinishRegistration(user, session, response)
}
