package jwt

import (
	"context"
	"errors"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"time"
)

type Port interface {
	GetAccessTime() time.Time
	GetRefreshTime() time.Time
	GenPair(ctx context.Context, uid uuid.UUID, roles []md.Role) (string, string, error)
	NewToken(ctx context.Context, uid uuid.UUID, roles []md.Role, d time.Duration) (string, error)
	ParseClaims(ctx context.Context, tokenStr string) (Claims, error)
}

const AccessTokenDuration = time.Minute * 30
const RefreshTokenDuration = time.Hour * 24 * 7

var ErrWhileCreatingToken = errors.New("error while creating token")
var ErrUnexpectedSignMethod = errors.New("unexpected signing method")
var ErrInvalidToken = errors.New("invalid token")

type Core struct {
	secret []byte
	issuer string
}

type Claims struct {
	UID   uuid.UUID `json:"uid"`
	Roles []md.Role `json:"roles"`
	jwt.RegisteredClaims
}

func New(conf config.Config) *Core {
	return &Core{secret: []byte(conf.Auth.JWT.Secret), issuer: conf.Auth.JWT.Issuer}
}

func (c *Core) GetAccessTime() time.Time {
	return time.Now().Add(AccessTokenDuration)
}

func (c *Core) GetRefreshTime() time.Time {
	return time.Now().Add(RefreshTokenDuration)
}

func (c *Core) GenPair(ctx context.Context, uid uuid.UUID, roles []md.Role) (string, string, error) {
	const op = "auth.GenPair.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	access, err := c.NewToken(ctx, uid, roles, AccessTokenDuration)
	if err != nil {
		zap.L().Error(
			"Failed to generate token pair",
			zap.String("uid", uid.String()),
			zap.Any("roles", roles),
			zap.Error(err),
		)
		return "", "", err
	}

	refresh, err := c.NewToken(ctx, uid, roles, RefreshTokenDuration)
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

func (c *Core) NewToken(ctx context.Context, uid uuid.UUID, roles []md.Role, d time.Duration) (string, error) {
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
				Issuer:    c.issuer,
			},
		},
	).SignedString(c.secret)

	if err != nil {
		zap.L().Error(
			ErrWhileCreatingToken.Error(),
			zap.Error(err),
		)
		return "", ErrWhileCreatingToken
	}

	return signed, nil
}

func (c *Core) ParseClaims(ctx context.Context, tokenStr string) (Claims, error) {
	const op = "auth.ParseClaims.jwt"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	claims := Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr, &claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrUnexpectedSignMethod
			}
			return c.secret, nil
		},
	)

	if err != nil {
		zap.L().Error(
			"Failed to parse claims",
			zap.String("op", op),
			zap.Any("token", tokenStr),
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
