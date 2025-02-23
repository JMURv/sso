package auth

import (
	"context"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

type Claims struct {
	UID   uuid.UUID       `json:"uid"`
	Roles []md.Permission `json:"roles"`
	jwt.RegisteredClaims
}

type Auth struct {
	secret        []byte
	refreshSecret []byte
}

func New(conf *config.AuthConfig) {
	Au = &Auth{
		secret:        []byte(conf.Secret),
		refreshSecret: []byte(conf.RefreshSecret),
	}
}

func (a *Auth) Hash(val string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.DefaultCost)
	return string(bytes), err
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

	token, err := jwt.Parse(
		tokenStr, func(token *jwt.Token) (any, error) {
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

	claims, ok := token.Claims.(*Claims)
	if !ok {
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
