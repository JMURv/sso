package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	md "github.com/JMURv/sso/internal/models"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

var ErrAuthHeaderIsMissing = errors.New("authorization header is missing")
var ErrInvalidTokenFormat = errors.New("invalid token format")

func Auth(au auth.Core) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				header := r.Header.Get("Authorization")
				if header == "" {
					utils.ErrResponse(w, http.StatusForbidden, ErrAuthHeaderIsMissing)
					return
				}

				token := strings.TrimPrefix(header, "Bearer ")
				if token == header {
					utils.ErrResponse(w, http.StatusForbidden, ErrInvalidTokenFormat)
					return
				}

				claims, err := au.ParseClaims(r.Context(), token)
				if err != nil {
					utils.ErrResponse(w, http.StatusForbidden, err)
					return
				}

				ctx := context.WithValue(r.Context(), "uid", claims.UID)
				ctx = context.WithValue(ctx, "roles", claims.Roles)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

var ErrNotAuthorized = errors.New("not authorized")

func CheckRights(c ctrl.AppCtrl) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				uid, ok := r.Context().Value("uid").(uuid.UUID)
				if !ok {
					zap.L().Debug(
						hdl.ErrFailedToParseUUID.Error(),
						zap.Any("uid", r.Context().Value("uid")),
					)
					utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
				}

				roles, ok := r.Context().Value("roles").([]md.Role)
				if !ok {
					zap.L().Debug(
						hdl.ErrFailedToParseRoles.Error(),
						zap.Any("uid", r.Context().Value("uid")),
					)
					utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseRoles)
				}

				for i := 0; i < len(roles); i++ {
					if roles[i].Name == "admin" {
						next.ServeHTTP(w, r)
						return
					}
				}

				dID := chi.URLParam(r, "id")
				if dID == "" {
					zap.L().Debug(
						hdl.ErrToRetrievePathArg.Error(),
						zap.String("path", r.URL.Path),
					)
					utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrToRetrievePathArg)
					return
				}

				device, err := c.GetDeviceByID(r.Context(), dID)
				if err != nil {
					utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
					return
				}

				if device.UserID == uid {
					next.ServeHTTP(w, r)
					return
				}

				utils.ErrResponse(w, http.StatusForbidden, ErrNotAuthorized)
			},
		)
	}
}

var ErrIPIsIncorrect = errors.New("ip is incorrect")
var ErrUAIsIncorrect = errors.New("user agent is incorrect")

func Device(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if ip == "" {
				utils.ErrResponse(w, http.StatusForbidden, ErrIPIsIncorrect)
			}

			splitIP := strings.Split(ip, ".")
			if len(splitIP) != 4 {
				utils.ErrResponse(w, http.StatusForbidden, ErrIPIsIncorrect)
			}

			ua := r.UserAgent()
			if ua == "" {
				utils.ErrResponse(w, http.StatusForbidden, ErrUAIsIncorrect)
			}
			ctx := context.WithValue(r.Context(), "ip", ip)
			ctx = context.WithValue(ctx, "ua", ua)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Prometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s := time.Now()
			op := fmt.Sprintf("%s %s", r.Method, r.RequestURI)
			span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
			defer span.Finish()

			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r.WithContext(ctx))
			metrics.ObserveRequest(time.Since(s), lrw.statusCode, op)
		},
	)
}

func OT(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			op := fmt.Sprintf("%s %s", r.Method, r.RequestURI)
			span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
			defer span.Finish()

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
