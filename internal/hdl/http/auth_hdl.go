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

func RegisterAuthRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc("/api/auth/jwt", mid.Apply(
		h.authenticate,
		mid.AllowedMethods(http.MethodPost),
		mid.Device,
	))

	mux.HandleFunc("/api/auth/jwt/parse", mid.Apply(
		h.parseClaims,
		mid.AllowedMethods(http.MethodPost),
	))

	mux.HandleFunc("/api/auth/jwt/refresh", mid.Apply(
		h.refresh,
		mid.AllowedMethods(http.MethodPost),
		mid.Device,
	))

	mux.HandleFunc("/api/auth/email/send", mid.Apply(
		h.sendLoginCode,
		mid.AllowedMethods(http.MethodPost),
	))

	mux.HandleFunc("/api/auth/email/check", mid.Apply(
		h.checkLoginCode,
		mid.AllowedMethods(http.MethodPost),
		mid.Device,
	))

	mux.HandleFunc("/api/auth/recovery/send", mid.Apply(
		h.sendForgotPasswordEmail,
		mid.AllowedMethods(http.MethodPost),
	))
	mux.HandleFunc("/api/auth/recovery/check", mid.Apply(
		h.checkForgotPasswordEmail,
		mid.AllowedMethods(http.MethodPost),
	))

	mux.HandleFunc("/api/auth/logout", mid.Apply(
		h.logout,
		mid.AllowedMethods(http.MethodPost),
		mid.Auth(au),
	))
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.EmailAndPasswordRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.Authenticate(r.Context(), &d, req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, auth.ErrInvalidCredentials) {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.RefreshRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.Refresh(r.Context(), &d, req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, auth.ErrTokenRevoked) {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) parseClaims(w http.ResponseWriter, r *http.Request) {
	req := &dto.TokenRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.ParseClaims(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error(
			hdl.ErrFailedToGetUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	err := h.ctrl.Logout(r.Context(), uid)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
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

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) sendForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	req := &dto.SendForgotPasswordEmail{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	err := h.ctrl.SendForgotPasswordEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) checkForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	req := &dto.CheckForgotPasswordEmailRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	err := h.ctrl.CheckForgotPasswordEmail(r.Context(), req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, ctrl.ErrCodeIsNotValid) {
			utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrInternal)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) sendLoginCode(w http.ResponseWriter, r *http.Request) {
	req := &dto.LoginCodeRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	err := h.ctrl.SendLoginCode(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) checkLoginCode(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.CheckLoginCodeRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.CheckLoginCode(r.Context(), &d, req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.SuccessResponse(w, http.StatusOK, res)
}
