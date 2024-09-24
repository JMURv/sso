package auth

import (
	"errors"
	"github.com/JMURv/sso/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const AccessTokenDuration = time.Hour * 72
const RefreshTokenDuration = time.Hour * 24 * 30

var ErrInvalidToken = errors.New("invalid token")

type Auth struct {
	secret string
}

func New(secret string) *Auth {
	return &Auth{secret: secret}
}

func (a *Auth) NewToken(u *model.User, d time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = u.ID
	claims["email"] = u.Email
	claims["exp"] = time.Now().Add(d).Unix()

	tokenStr, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (a *Auth) VerifyToken(tokenStr string) (map[string]any, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(a.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, ErrInvalidToken
	}
}
