package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/JMURv/sso/internal/hdl/validation"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

func RegisterAuthRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/auth/jwt", mid.Apply(h.authenticate, mid.Device))
	mux.HandleFunc("/api/auth/jwt/parse", h.parseClaims)
	mux.HandleFunc("/api/auth/jwt/refresh", mid.Apply(h.refresh, mid.Device))

	mux.HandleFunc("/api/auth/oauth2/{provider}/start", h.startOAuth2)
	mux.HandleFunc("/api/auth/oauth2/{provider}/callback", mid.Apply(h.handleOAuth2Callback, mid.Device))

	mux.HandleFunc("/api/auth/oidc/{provider}/start", h.startOIDC)
	mux.HandleFunc("/api/auth/oidc/{provider}/callback", mid.Apply(h.handleOIDCCallback, mid.Device))

	mux.HandleFunc("/api/auth/webauthn/register/start", mid.Apply(h.webauthnRegistrationStart, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/register/finish", mid.Apply(h.webauthnRegistrationFinish, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/login/start", h.webauthnLoginStart)
	mux.HandleFunc("/api/auth/webauthn/login/finish", h.webauthnLoginFinish)

	mux.HandleFunc("/api/auth/email/send", h.sendLoginCode)
	mux.HandleFunc("/api/auth/email/check", mid.Apply(h.checkLoginCode, mid.Device))

	mux.HandleFunc("/api/auth/recovery/send", h.sendForgotPasswordEmail)
	mux.HandleFunc("/api/auth/recovery/check", h.checkForgotPasswordEmail)

	mux.HandleFunc("/api/auth/logout", mid.Apply(h.logout, mid.Auth))
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) {
	const op = "auth.authenticate.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.EmailAndPasswordRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.LoginAndPasswordRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.Authenticate(ctx, &d, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil && errors.Is(err, auth.ErrInvalidCredentials) {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, res)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	const op = "auth.refresh.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.RefreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.RefreshRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.Refresh(ctx, &d, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil && errors.Is(err, auth.ErrTokenRevoked) {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, res)
}

func (h *Handler) parseClaims(w http.ResponseWriter, r *http.Request) {
	const op = "auth.parseClaims.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	req := &dto.TokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.TokenRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.ParseClaims(ctx, req.Token)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, c, res)

}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	const op = "auth.logout.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToGetUUID.Error(),
			zap.String("op", op),
			zap.Any("uid", ctx.Value("uid")),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	err := h.ctrl.Logout(ctx, uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	http.SetCookie(
		w, &http.Cookie{
			Name:     "access",
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	http.SetCookie(
		w, &http.Cookie{
			Name:     "refresh",
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	utils.StatusResponse(w, c)
}

func (h *Handler) startOAuth2(w http.ResponseWriter, r *http.Request) {
	const op = "auth.startOAuth2.hdl"
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

func (h *Handler) sendForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	const op = "auth.sendForgotPasswordEmail.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	req := &dto.SendForgotPasswordEmail{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.SendForgotPasswordEmail(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	err := h.ctrl.SendForgotPasswordEmail(ctx, req.Email)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, c)
}

func (h *Handler) checkForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	const op = "auth.checkForgotPasswordEmail.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	req := &dto.CheckForgotPasswordEmailRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.CheckForgotPasswordEmailRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	err := h.ctrl.CheckForgotPasswordEmail(ctx, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil && errors.Is(err, ctrl.ErrCodeIsNotValid) {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, c)
}

func (h *Handler) sendLoginCode(w http.ResponseWriter, r *http.Request) {
	const op = "auth.sendLoginCode.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	req := &dto.LoginCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.LoginCodeRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	err := h.ctrl.SendLoginCode(ctx, req.Email, req.Password)
	if err != nil && errors.Is(err, auth.ErrInvalidCredentials) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.StatusResponse(w, c)
}

func (h *Handler) checkLoginCode(w http.ResponseWriter, r *http.Request) {
	const op = "auth.checkLoginCode.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.CheckLoginCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.String("op", op),
			zap.Any("req", req),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.CheckLoginCodeRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.CheckLoginCode(ctx, &d, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
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
