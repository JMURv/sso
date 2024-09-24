package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	controller "github.com/JMURv/sso/internal/controller"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	srv  *http.Server
	ctrl *controller.Controller
	auth *auth.Auth
}

func New(auth *auth.Auth, ctrl *controller.Controller) *Handler {
	return &Handler{
		auth: auth,
		ctrl: ctrl,
	}
}

func (h *Handler) Start(port int) {
	r := mux.NewRouter()
	r.Use(h.tracingMiddleware)

	r.HandleFunc("/api/health-check", h.healthCheck).Methods(http.MethodGet)

	RegisterAuthRoutes(r, h)
	RegisterUserRoutes(r, h)

	h.srv = &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%v", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := h.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		zap.L().Debug("Server error", zap.Error(err))
	}
}

func (h *Handler) Close() error {
	if err := h.srv.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

func middlewareFunc(h http.HandlerFunc, middleware ...func(http.Handler) http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler = h
		for _, m := range middleware {
			handler = m(handler)
		}
		handler.ServeHTTP(w, r)
	}
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrResponse(w, http.StatusUnauthorized, errors.New("authorization header is missing"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			utils.ErrResponse(w, http.StatusUnauthorized, errors.New("invalid token format"))
			return
		}

		claims, err := h.auth.VerifyToken(tokenStr)
		if err != nil {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		}

		ctx := context.WithValue(r.Context(), "uid", claims["uid"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := opentracing.GlobalTracer().StartSpan(
			fmt.Sprintf("%s %s", r.Method, r.URL),
		)
		defer span.Finish()

		zap.L().Info("Request", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "OK")
}
