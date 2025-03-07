package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	wa "github.com/JMURv/sso/internal/auth/webauthn"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/JMURv/sso/internal/hdl/validation"
	md "github.com/JMURv/sso/internal/models"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func RegisterWebAuthnRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/auth/webauthn/register/start", mid.Apply(h.registrationStart, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/register/finish", mid.Apply(h.registrationFinish, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/login/start", h.loginStart)
	mux.HandleFunc("/api/auth/webauthn/login/finish", h.loginFinish)
}

func (h *Handler) registrationStart(w http.ResponseWriter, r *http.Request) {
	const op = "webauthn.registrationStart.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if !ok {
		c = http.StatusUnauthorized
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op),
			zap.Error(hdl.ErrFailedToParseUUID),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	user, err := h.ctrl.GetUserForWA(ctx, uid)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrInternal.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, err)
		return
	}

	options, sessionData, err := auth.Au.Wa.BeginRegistration(user)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrInternal.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, err)
		return
	}

	if err := h.ctrl.StoreWASession(ctx, wa.Register, uid, sessionData); err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrInternal.Error(),
			zap.String("op", op),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, options)
}

func (h *Handler) registrationFinish(w http.ResponseWriter, r *http.Request) {
	const op = "webauthn.registrationFinish.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrInternal)
		return
	}

	user, err := h.ctrl.GetUserForWA(ctx, uid)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	sessionData, err := h.ctrl.GetWASession(ctx, wa.Register, uid)
	if err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, errors.New("invalid session"))
		return
	}

	credential, err := auth.Au.Wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		zap.L().Error("failed to finish registration", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, http.StatusBadRequest, errors.New("invalid registration"))
		return
	}

	if err := h.ctrl.StoreWACredential(ctx, uid, credential); err != nil {
		zap.L().Error("failed to store credential", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, c)
}

func (h *Handler) loginStart(w http.ResponseWriter, r *http.Request) {
	const op = "webauthn.loginStart.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	var req dto.LoginStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.V.Struct(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	user, err := h.ctrl.GetUserByEmailForWA(ctx, req.Email)
	if err != nil {
		utils.ErrResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	}

	options, sessionData, err := auth.Au.Wa.BeginLogin(user)
	if err != nil {
		zap.L().Error("failed to begin login", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	if err := h.ctrl.StoreWASession(ctx, wa.Login, user.ID, sessionData); err != nil {
		zap.L().Error("failed to store session", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, c, options)
}

func (h *Handler) loginFinish(w http.ResponseWriter, r *http.Request) {
	const op = "webauthn.loginFinish.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, hdl.ErrNoDeviceInfo)
		return
	}

	var req dto.LoginFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.V.Struct(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	user, err := h.ctrl.GetUserByEmailForWA(ctx, req.Email)
	if err != nil {
		utils.ErrResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	}

	sessionData, err := h.ctrl.GetWASession(ctx, wa.Login, user.ID)
	if err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, errors.New("invalid session"))
		return
	}

	_, err = auth.Au.Wa.FinishLogin(user, *sessionData, r)
	if err != nil {
		zap.L().Error("failed to finish login", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, http.StatusUnauthorized, errors.New("authentication failed"))
		return
	}

	res, err := h.ctrl.GenPair(ctx, &d, user.ID, []md.Permission{})
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, c, res)
}
