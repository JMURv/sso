package http

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Handler struct {
	srv  *http.Server
	ctrl ctrl.AppCtrl
	au   auth.Core
}

func New(ctrl ctrl.AppCtrl, au auth.Core) *Handler {
	return &Handler{
		ctrl: ctrl,
		au:   au,
	}
}

func (h *Handler) Start(port int) {
	mux := http.NewServeMux()

	RegisterAuthRoutes(mux, h.au, h)
	RegisterOAuth2Routes(mux, h)
	RegisterOIDCRoutes(mux, h)
	RegisterWebAuthnRoutes(mux, h.au, h)

	RegisterUserRoutes(mux, h.au, h)
	RegisterPermRoutes(mux, h.au, h)
	RegisterDeviceRoutes(mux, h.au, h)
	mux.HandleFunc(
		"/health", func(w http.ResponseWriter, r *http.Request) {
			utils.SuccessResponse(w, http.StatusOK, "OK")
		},
	)

	handler := mid.LogTraceMetrics(mux)
	handler = mid.RecoverPanic(handler)
	h.srv = &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf(":%v", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	zap.L().Info(
		"Starting HTTP server",
		zap.String("addr", h.srv.Addr),
	)

	err := h.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		zap.L().Error("Server error", zap.Error(err))
	}
}

func (h *Handler) Close(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}
