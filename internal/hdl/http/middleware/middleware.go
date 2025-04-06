package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"slices"
	"strings"
	"time"
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

func Auth(au auth.Core) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				header := r.Header.Get("Authorization")
				if header == "" {
					utils.ErrResponse(w, http.StatusUnauthorized, ErrAuthHeaderIsMissing)
					return
				}

				token := strings.TrimPrefix(header, "Bearer ")
				if token == header {
					utils.ErrResponse(w, http.StatusUnauthorized, ErrInvalidTokenFormat)
					return
				}

				claims, err := au.ParseClaims(r.Context(), token)
				if err != nil {
					utils.ErrResponse(w, http.StatusUnauthorized, err)
					return
				}

				ctx := context.WithValue(r.Context(), "uid", claims.UID)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func Device(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			ip := r.Header.Get("X-Forwarded-For")

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

var ErrMethodNotAllowed = errors.New("method not allowed")

func AllowedMethods(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if ok := slices.Contains(methods, r.Method); !ok {
					utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
					return
				}
				next.ServeHTTP(w, r)
			},
		)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LogTraceMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s := time.Now()
			op := fmt.Sprintf("%s %s", r.Method, r.RequestURI)
			span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
			defer span.Finish()

			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r.WithContext(ctx))
			metrics.ObserveRequest(time.Since(s), lrw.statusCode, op)

			zap.L().Info(
				"<--",
				zap.String("method", r.Method),
				zap.Int("status", lrw.statusCode),
				zap.Any("duration", time.Since(s)),
				zap.String("uri", r.RequestURI),
			)
		},
	)
}
