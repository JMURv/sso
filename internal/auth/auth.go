package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	HashSHA256(val string) (string, error)
	ComparePasswords(hashed, pswd []byte) error
	NewToken(ctx context.Context, uid uuid.UUID, perms []md.Permission, d time.Duration) (string, error)
	ParseClaims(ctx context.Context, tokenStr string) (Claims, error)
	GenPair(ctx context.Context, uid uuid.UUID, perms []md.Permission) (string, string, error)
}

type Claims struct {
	UID   uuid.UUID       `json:"uid"`
	Roles []md.Permission `json:"roles"`
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *Auth) HashSHA256(val string) (string, error) {
	shaHash := sha256.Sum256([]byte(val))
	hexHash := hex.EncodeToString(shaHash[:])
	bytes, err := bcrypt.GenerateFromPassword([]byte(hexHash), bcrypt.DefaultCost)
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
		return claims, err
	}

	if !token.Valid {
		return claims, ErrInvalidToken
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
