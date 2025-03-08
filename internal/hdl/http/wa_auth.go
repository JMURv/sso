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
	"time"
)

func RegisterWebAuthnRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc("/api/auth/webauthn/register/start", mid.Apply(h.registrationStart, mid.Auth(au)))
	mux.HandleFunc("/api/auth/webauthn/register/finish", mid.Apply(h.registrationFinish, mid.Auth(au)))
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

	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op),
			zap.Any("uid", ctx.Value("uid")),
			zap.Error(hdl.ErrFailedToParseUUID),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	res, err := h.ctrl.StartRegistration(ctx, uid)
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

	utils.SuccessResponse(w, c, res)
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
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op),
			zap.Any("uid", ctx.Value("uid")),
			zap.Error(hdl.ErrFailedToParseUUID),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	err := h.ctrl.FinishRegistration(ctx, uid, r)
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

	res, err := h.ctrl.BeginLogin(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			c = http.StatusNotFound
			utils.ErrResponse(w, c, err)
			return
		}
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, c, res)
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

	res, err := h.ctrl.FinishLogin(ctx, req.Email, d, r)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, c, res)
}
