package http

import (
	"errors"
	"net/http"

	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/auth/captcha"
	_ "github.com/JMURv/sso/internal/auth/jwt"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) RegisterAuthRoutes() {
	h.router.With(mid.Device).Post("/auth/jwt", h.authenticate)
	h.router.Post("/auth/jwt/parse", h.parseClaims)
	h.router.With(mid.Device).Post("/auth/jwt/refresh", h.refresh)
	h.router.With(mid.Device).Post("/auth/email/send", h.sendLoginCode)
	h.router.With(mid.Device).Post("/auth/email/check", h.checkLoginCode)
	h.router.Post("/auth/recovery/send", h.sendForgotPasswordEmail)
	h.router.Post("/auth/recovery/check", h.checkForgotPasswordEmail)
	h.router.With(mid.Auth(h.au)).Post("/auth/logout", h.logout)
}

// authenticate godoc
//
//	@Summary		Authenticate using email & password
//	@Description	Verify reCAPTCHA, then authenticate and set JWT cookies
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			X-Real-IP	header		string						true	"Client real IP address"
//	@Param			User-Agent	header		string						true	"Client User-Agent"
//	@Param			body		body		dto.EmailAndPasswordRequest	true	"email, password, reCAPTCHA token"
//	@Success		200			{object}	dto.TokenPair
//	@Failure		400			{object}	utils.ErrorsResponse	"missing device info or bad payload"
//	@Failure		401			{object}	utils.ErrorsResponse	"invalid credentials or reCAPTCHA"
//	@Failure		404			{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/jwt [post]
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

	valid, err := h.au.VerifyRecaptcha(req.Token, captcha.PassAuth)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, captcha.ErrVerificationFailed)
		return
	}

	if !valid {
		utils.ErrResponse(w, http.StatusUnauthorized, captcha.ErrValidationFailed)
		return
	}

	res, err := h.ctrl.Authenticate(r.Context(), &d, req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.StatusResponse(w, http.StatusOK)
}

// refresh godoc
//
//	@Summary		Refresh JWT tokens
//	@Description	Validate device header and refresh tokens, reset cookies
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			X-Real-IP	header		string				true	"Client real IP address"
//	@Param			User-Agent	header		string				true	"Client User-Agent"
//	@Param			body		body		dto.RefreshRequest	true	"refresh_token"
//	@Success		200			{object}	dto.TokenPair
//	@Failure		400			{object}	utils.ErrorsResponse	"missing device info or bad payload"
//	@Failure		401			{object}	utils.ErrorsResponse	"token revoked or invalid"
//	@Failure		404			{object}	utils.ErrorsResponse	"session not found"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/jwt/refresh [post]
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	cookie, err := r.Cookie(config.RefreshCookieName)
	if err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	res, err := h.ctrl.Refresh(
		r.Context(), &d, &dto.RefreshRequest{
			Refresh: cookie.Value,
		},
	)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else if errors.Is(err, auth.ErrTokenRevoked) {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.StatusResponse(w, http.StatusOK)
}

// parseClaims godoc
//
//	@Summary		Parse JWT claims
//	@Description	Decode a token without requiring device header
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.TokenRequest	true	"jwt token"
//	@Success		200		{object}	jwt.Claims
//	@Failure		404		{object}	utils.ErrorsResponse	"token not found"
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/jwt/parse [post]
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

// logout godoc
//
//	@Summary		Logout user
//	@Description	Revoke refresh token, clear JWT cookies
//	@Tags			Authentication
//	@Produce		json
//	@Param			Authorization	header		string					true	"Authorization token"
//	@Success		200				{object}	nil						"OK"
//	@Failure		404				{object}	utils.ErrorsResponse	"session not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/logout [post]
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

// sendLoginCode godoc
//
//	@Summary		Send login code via email
//	@Description	Verify reCAPTCHA, then send a one-time code to the user’s email. May return tokens if password also valid.
//	@Tags			EmailAuth
//	@Accept			json
//	@Produce		json
//	@Param			X-Real-IP	header		string					true	"Client real IP address"
//	@Param			User-Agent	header		string					true	"Client User-Agent"
//	@Param			body		body		dto.LoginCodeRequest	true	"email, password, reCAPTCHA token"
//	@Success		200			{object}	dto.TokenPair
//	@Failure		400			{object}	utils.ErrorsResponse	"missing device info or bad payload"
//	@Failure		401			{object}	utils.ErrorsResponse	"invalid credentials or reCAPTCHA"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/email/send [post]
func (h *Handler) sendLoginCode(w http.ResponseWriter, r *http.Request) {
	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	req := &dto.LoginCodeRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	valid, err := h.au.VerifyRecaptcha(req.Token, captcha.EmailAuth)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, captcha.ErrVerificationFailed)
		return
	}

	if !valid {
		utils.ErrResponse(w, http.StatusUnauthorized, captcha.ErrValidationFailed)
		return
	}

	res, err := h.ctrl.SendLoginCode(r.Context(), &d, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		} else {
			utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
			return
		}
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	utils.StatusResponse(w, http.StatusOK)
}

// checkLoginCode godoc
//
//	@Summary		Check email login code
//	@Description	Exchange a valid email code for JWT tokens
//	@Tags			EmailAuth
//	@Accept			json
//	@Produce		json
//	@Param			X-Real-IP	header		string						true	"Client real IP address"
//	@Param			User-Agent	header		string						true	"Client User-Agent"
//	@Param			body		body		dto.CheckLoginCodeRequest	true	"code, reCAPTCHA token"
//	@Success		200			{object}	dto.TokenPair
//	@Failure		400			{object}	utils.ErrorsResponse	"missing device info or bad payload"
//	@Failure		404			{object}	utils.ErrorsResponse	"code not found"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/email/check [post]
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
	utils.StatusResponse(w, http.StatusOK)
}

// sendForgotPasswordEmail godoc
//
//	@Summary		Send forgot‐password email
//	@Description	Verify reCAPTCHA and send recovery email
//	@Tags			PasswordRecovery
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.SendForgotPasswordEmail	true	"email, reCAPTCHA token"
//	@Success		200		{object}	nil							"OK"
//	@Failure		404		{object}	utils.ErrorsResponse		"email not found"
//	@Failure		500		{object}	utils.ErrorsResponse		"internal error"
//	@Router			/auth/recovery/send [post]
func (h *Handler) sendForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	req := &dto.SendForgotPasswordEmail{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	valid, err := h.au.VerifyRecaptcha(req.Token, captcha.ForgotPass)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, captcha.ErrVerificationFailed)
		return
	}

	if !valid {
		utils.ErrResponse(w, http.StatusUnauthorized, captcha.ErrValidationFailed)
		return
	}

	err = h.ctrl.SendForgotPasswordEmail(r.Context(), req.Email)
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

// checkForgotPasswordEmail godoc
//
//	@Summary		Check forgot‐password code
//	@Description	Validate a recovery code
//	@Tags			PasswordRecovery
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CheckForgotPasswordEmailRequest	true	"code"
//	@Success		200		{object}	nil									"OK"
//	@Failure		401		{object}	utils.ErrorsResponse				"invalid code"
//	@Failure		404		{object}	utils.ErrorsResponse				"code not found"
//	@Failure		500		{object}	utils.ErrorsResponse				"internal error"
//	@Router			/auth/recovery/check [post]
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
		}
		if errors.Is(err, ctrl.ErrCodeIsNotValid) {
			utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrInternal)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}
