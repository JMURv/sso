package middleware

import (
	"context"
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

var ErrAuthHeaderIsMissing = errors.New("authorization header is missing")
var ErrInvalidTokenFormat = errors.New("invalid token format")

func Apply(h http.HandlerFunc, middleware ...func(http.Handler) http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler = h
		for _, m := range middleware {
			handler = m(handler)
		}
		handler.ServeHTTP(w, r)
	}
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ErrResponse(w, http.StatusUnauthorized, ErrAuthHeaderIsMissing)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				utils.ErrResponse(w, http.StatusUnauthorized, ErrInvalidTokenFormat)
				return
			}

			claims, err := auth.Au.ParseClaims(r.Context(), tokenStr)
			if err != nil {
				utils.ErrResponse(w, http.StatusUnauthorized, err)
				return
			}

			ctx := context.WithValue(r.Context(), "uid", claims.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func Device(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = strings.Split(forwarded, ",")[0]
			}

			ctx := context.WithValue(r.Context(), "ip", ip)
			ctx = context.WithValue(ctx, "ua", ua)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					zap.L().Error("panic", zap.Any("err", err))
					utils.ErrResponse(w, http.StatusInternalServerError, errors.New("internal error"))
				}
			}()
			next.ServeHTTP(w, r)
		},
	)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			zap.L().Info(
				"Request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
			)
			next.ServeHTTP(w, r)
		},
	)
}
