package http

import (
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strings"
	"time"
)

func RegisterOIDCRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/auth/oidc/{provider}/start", h.startOIDC)
	mux.HandleFunc("/api/auth/oidc/{provider}/callback", mid.Apply(h.handleOIDCCallback, mid.Device))
}

func (h *Handler) startOIDC(w http.ResponseWriter, r *http.Request) {
	const op = "auth.startOIDC.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, ErrInvalidURL)
		return
	}
	provider := parts[4]

	res, err := h.ctrl.GetOIDCAuthURL(ctx, provider)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handleOIDCCallback.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, ErrInvalidURL)
		return
	}
	provider := parts[4]
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrNoDeviceInfo)
		return
	}

	res, err := h.ctrl.HandleOIDCCallback(ctx, &d, provider, code, state)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	http.SetCookie(
		w, &http.Cookie{
			Name:     "access",
			Value:    res.Access,
			Expires:  time.Now().Add(auth.AccessTokenDuration),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	http.SetCookie(
		w, &http.Cookie{
			Name:     "refresh",
			Value:    res.Refresh,
			Expires:  time.Now().Add(auth.RefreshTokenDuration),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	utils.SuccessResponse(w, c, res)
}
