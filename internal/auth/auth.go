package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	pr_oauth2 "github.com/JMURv/sso/internal/auth/providers/oauth2/google"
	oidc "github.com/JMURv/sso/internal/auth/providers/oidc/google"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	md "github.com/JMURv/sso/internal/models"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"strconv"
	"strings"
	"time"
)

const issuser = "SSO"
const AccessTokenDuration = time.Minute * 30
const RefreshTokenDuration = time.Hour * 24 * 7

var Au *Auth

type AuthService interface {
	NewToken(ctx context.Context, uid uuid.UUID) (string, error)
	VerifyToken(ctx context.Context, tokenStr string) (map[string]any, error)
	Hash(string) (string, error)
	ComparePasswords(hashed, pswd []byte) error
}

type OAuth2Provider interface {
	GetName() string
	GetConfig() *oauth2.Config
	GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error)
}

type OIDCProvider interface {
	GetName() string
	GetConfig() *oauth2.Config
	GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error)
}

type Claims struct {
	UID   uuid.UUID       `json:"uid"`
	Roles []md.Permission `json:"roles"`
	jwt.RegisteredClaims
}

type Auth struct {
	secret          []byte
	OAuth2Providers struct {
		Google OAuth2Provider
	}
	OIDCProviders struct {
		Google OIDCProvider
	}
}

func New(conf config.Config) {
	Au = &Auth{
		secret: []byte(conf.Auth.Secret),
		OAuth2Providers: struct {
			Google OAuth2Provider
		}{
			Google: pr_oauth2.NewGoogleOAuth2(conf),
		},
		OIDCProviders: struct {
			Google OIDCProvider
		}{
			Google: oidc.NewGoogleOIDC(conf),
		},
	}
}

func (a *Auth) Hash(val string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *Auth) HashSHA256(val string) (string, error) {
	shaHash := sha256.Sum256([]byte(val))
	hexHash := hex.EncodeToString(shaHash[:])
	bytes, err := bcrypt.GenerateFromPassword([]byte(hexHash), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *Auth) GenerateSignedState() (string, error) {
	rawState := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	data := rawState + "|" + timestamp

	h := hmac.New(sha256.New, a.secret)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	signedState := base64.URLEncoding.EncodeToString([]byte(data)) + "." + base64.URLEncoding.EncodeToString(signature)
	return signedState, nil
}

func (a *Auth) ValidateSignedState(signedState string, maxAge time.Duration) (bool, error) {
	parts := strings.Split(signedState, ".")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid state format")
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	expectedSignature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	h := hmac.New(sha256.New, a.secret)
	h.Write(data)
	actualSignature := h.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		return false, fmt.Errorf("invalid signature")
	}

	dataParts := strings.Split(string(data), "|")
	if len(dataParts) != 2 {
		return false, fmt.Errorf("invalid data format")
	}

	timestamp, err := strconv.ParseInt(dataParts[1], 10, 64)
	if err != nil {
		return false, err
	}

	if time.Since(time.Unix(timestamp, 0)) > maxAge {
		return false, fmt.Errorf("state expired")
	}

	return true, nil
}

func (a *Auth) ComparePasswords(hashed, pswd []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashed, pswd); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (a *Auth) NewToken(ctx context.Context, uid uuid.UUID, perms []md.Permission, d time.Duration) (string, error) {
	const op = "auth.NewToken.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	signed, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			UID:   uid,
			Roles: perms,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    issuser,
			},
		},
	).SignedString(a.secret)

	if err != nil {
		zap.L().Error(
			ErrWhileCreatingToken.Error(),
			zap.Error(err),
		)
		return "", ErrWhileCreatingToken
	}

	return signed, nil
}

func (a *Auth) ParseClaims(ctx context.Context, tokenStr string) (*Claims, error) {
	const op = "auth.ParseClaims.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr, claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrUnexpectedSignMethod
			}
			return a.secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (a *Auth) GenPair(ctx context.Context, uid uuid.UUID, perms []md.Permission) (string, string, error) {
	const op = "auth.GenPair.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	access, err := a.NewToken(ctx, uid, perms, AccessTokenDuration)
	if err != nil {
		return "", "", err
	}

	refresh, err := a.NewToken(ctx, uid, perms, RefreshTokenDuration)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (a *Auth) GetOAuth2Provider(ctx context.Context, provider string) (OAuth2Provider, error) {
	const op = "auth.GetOAuth2Provider.auth"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	switch provider {
	case "google":
		return a.OAuth2Providers.Google, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

type AuthFlow string

const (
	OIDC   AuthFlow = "oidc"
	OAuth2 AuthFlow = "oauth2"
)

type AuthProviders string

const (
	Google AuthProviders = "google"
	GitHub AuthProviders = "github"
)

func (a *Auth) GetProvider(ctx context.Context, provider AuthProviders, flow AuthFlow) (OIDCProvider, error) {
	const op = "auth.GetOAuth2Provider.auth"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	switch provider {
	case Google:
		if flow == OIDC {
			return a.OIDCProviders.Google, nil
		}
		return a.OAuth2Providers.Google, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
