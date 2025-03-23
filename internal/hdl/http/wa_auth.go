package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func RegisterWebAuthnRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc("/api/auth/webauthn/register/start", mid.Apply(
		h.registrationStart,
		mid.Auth(au),
	))

	mux.HandleFunc("/api/auth/webauthn/register/finish", mid.Apply(
		h.registrationFinish,
		mid.Auth(au),
	))

	mux.HandleFunc("/api/auth/webauthn/login/start", h.loginStart)
	mux.HandleFunc("/api/auth/webauthn/login/finish", mid.Apply(
		h.loginFinish,
		mid.Device,
	))
}

func (h *Handler) registrationStart(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
			zap.Error(hdl.ErrFailedToParseUUID),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	res, err := h.ctrl.StartRegistration(r.Context(), uid)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) registrationFinish(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
			zap.Error(hdl.ErrFailedToParseUUID),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	err := h.ctrl.FinishRegistration(r.Context(), uid, r)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) loginStart(w http.ResponseWriter, r *http.Request) {
	req := &dto.LoginStartRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.BeginLogin(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) loginFinish(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.LoginFinishRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.FinishLogin(r.Context(), req.Email, d, r)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, http.StatusOK, res)
}
