package http

import (
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

func RegisterOAuth2Routes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/auth/oauth2/{provider}/start", h.startOAuth2)
	mux.HandleFunc("/api/auth/oauth2/{provider}/callback", mid.Apply(h.handleOAuth2Callback, mid.Device))
}

func (h *Handler) startOAuth2(w http.ResponseWriter, r *http.Request) {
	const op = "auth.startOAuth2.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodGet {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, ErrInvalidURL)
		return
	}
	provider := parts[4]

	res, err := h.ctrl.GetOAuth2AuthURL(ctx, provider)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handleOAuth2Callback.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodGet {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

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

	res, err := h.ctrl.HandleOAuth2Callback(ctx, &d, provider, code, state)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrInternal.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, c, res)
}
