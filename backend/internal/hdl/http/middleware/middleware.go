package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	md "github.com/JMURv/sso/internal/models"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
)

func Auth(au auth.Core) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				access, err := r.Cookie(config.AccessCookieName)
				if err != nil {
					if errors.Is(err, http.ErrNoCookie) {
						utils.ErrResponse(w, http.StatusUnauthorized, err)
						return
					} else {
						zap.L().Error("failed to get access cookie", zap.Error(err))
						utils.ErrResponse(w, http.StatusInternalServerError, err)
						return
					}
				}

				claims, err := au.ParseClaims(r.Context(), access.Value)
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
					zap.L().Error(
						hdl.ErrFailedToParseUUID.Error(),
						zap.Any("uid", r.Context().Value("uid")),
					)
					utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
				}

				roles, ok := r.Context().Value("roles").([]md.Role)
				if !ok {
					zap.L().Error(
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
					zap.L().Error(
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
				return
			}

			ip = strings.Split(ip, ":")[0]
			splitIP := strings.Split(ip, ".")
			if len(splitIP) != 4 {
				utils.ErrResponse(w, http.StatusForbidden, ErrIPIsIncorrect)
				return
			}

			ua := r.UserAgent()
			if ua == "" {
				utils.ErrResponse(w, http.StatusForbidden, ErrUAIsIncorrect)
				return
			}

			zap.L().Debug("device info", zap.String("ip", ip), zap.String("ua", ua))
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

			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			metrics.ObserveRequest(time.Since(s), lrw.statusCode, op)
		},
	)
}

func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				lrw := NewLoggingResponseWriter(w)
				logger.Debug(
					"-->",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote", r.RemoteAddr),
				)

				next.ServeHTTP(lrw, r)

				logger.Info(
					"<--",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", lrw.statusCode),
					zap.Duration("duration", time.Since(start)),
					zap.String("remote", r.RemoteAddr),
				)
			},
		)
	}
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
