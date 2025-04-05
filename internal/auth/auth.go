package auth

import (
	"context"
	"github.com/JMURv/sso/internal/auth/providers"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const issuer = "SSO"
const AccessTokenDuration = time.Minute * 30
const RefreshTokenDuration = time.Hour * 24 * 7

type Core interface {
	Hash(val string) (string, error)
	ComparePasswords(hashed, pswd []byte) error
	GetRefreshTime() time.Time
	NewToken(ctx context.Context, uid uuid.UUID, perms []md.Role, d time.Duration) (string, error)
	ParseClaims(ctx context.Context, tokenStr string) (Claims, error)
	GenPair(ctx context.Context, uid uuid.UUID, perms []md.Role) (string, string, error)
}

type Claims struct {
	UID   uuid.UUID `json:"uid"`
	Roles []md.Role `json:"roles"`
	jwt.RegisteredClaims
}

type Auth struct {
	secret   []byte
	Provider *providers.Provider
	Wa       *wa.WAuthn
}

func New(conf config.Config) *Auth {
	return &Auth{
		secret:   []byte(conf.Auth.Secret),
		Provider: providers.New(conf),
		Wa:       wa.New(conf),
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

func (a *Auth) GetRefreshTime() time.Time {
	return time.Now().Add(RefreshTokenDuration)
}

func (a *Auth) ComparePasswords(hashed, pswd []byte) error {
	if err := bcrypt.CompareHashAndPassword(hashed, pswd); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (a *Auth) NewToken(ctx context.Context, uid uuid.UUID, roles []md.Role, d time.Duration) (string, error) {
	const op = "auth.NewToken.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	signed, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			UID:   uid,
			Roles: roles,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    issuer,
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

func (a *Auth) ParseClaims(ctx context.Context, tokenStr string) (Claims, error) {
	const op = "auth.ParseClaims.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims := Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr, &claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrUnexpectedSignMethod
			}
			return a.secret, nil
		},
	)

	if err != nil {
		zap.L().Error(
			"Failed to parse claims",
			zap.String("op", op),
			zap.String("token", tokenStr),
			zap.String("alg", token.Method.Alg()),
			zap.Error(err),
		)
		return claims, err
	}

	if !token.Valid {
		zap.L().Debug(
			"Token is invalid",
			zap.String("op", op),
			zap.String("token", tokenStr),
		)
		return claims, ErrInvalidToken
	}

	return claims, nil
}

func (a *Auth) GenPair(ctx context.Context, uid uuid.UUID, roles []md.Role) (string, string, error) {
	const op = "auth.GenPair.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	access, err := a.NewToken(ctx, uid, roles, AccessTokenDuration)
	if err != nil {
		zap.L().Error(
			"Failed to generate token pair",
			zap.String("uid", uid.String()),
			zap.Any("roles", roles),
			zap.Error(err),
		)
		return "", "", err
	}

	refresh, err := a.NewToken(ctx, uid, roles, RefreshTokenDuration)
	if err != nil {
		zap.L().Error(
			"Failed to generate token pair",
			zap.String("uid", uid.String()),
			zap.Any("roles", roles),
			zap.Error(err),
		)
		return "", "", err
	}

	return access, refresh, nil
}
