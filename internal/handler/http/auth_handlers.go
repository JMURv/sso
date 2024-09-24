package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/handler"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

func RegisterAuthRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/api/sso/send-login-code", h.sendLoginCode).Methods(http.MethodPost)
	r.HandleFunc("/api/sso/check-login-code", h.checkLoginCode).Methods(http.MethodPost)
	r.HandleFunc("/api/sso/check-email", h.checkEmail).Methods(http.MethodPost)
	r.HandleFunc("/api/sso/logout", middlewareFunc(h.logout, h.authMiddleware)).Methods(http.MethodPost)
	r.HandleFunc("/api/sso/recovery", h.sendForgotPasswordEmail).Methods(http.MethodPost)
	r.HandleFunc("/api/sso/recovery", h.checkForgotPasswordEmail).Methods(http.MethodPut)

	r.HandleFunc("/api/sso/support", middlewareFunc(h.sendSupportEmail, h.authMiddleware)).Methods(http.MethodPost)

	r.HandleFunc("/api/sso/me", middlewareFunc(h.me, h.authMiddleware)).Methods(http.MethodGet)
	r.HandleFunc("/api/sso/me", middlewareFunc(h.updateMe, h.authMiddleware)).Methods(http.MethodPut)
}

type sendSupportEmailRequest struct {
	Theme string `json:"theme"`
	Text  string `json:"text"`
}

func (h *Handler) sendSupportEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.sendSupportEmail.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(r.Context().Value("uid").(string))
	if err != nil {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, err)
		return
	}

	req := &sendSupportEmailRequest{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	err = h.ctrl.SendSupportEmail(r.Context(), uid, req.Theme, req.Text)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to send email", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

type checkForgotPasswordEmailRequest struct {
	Password string `json:"password"`
	Uidb64   string `json:"uidb64"`
	Token    string `json:"token"`
}

func (h *Handler) checkForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkForgotPasswordEmail.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	req := &checkForgotPasswordEmailRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	uidb64, err := uuid.Parse(req.Uidb64)
	if err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	intToken, err := strconv.Atoi(req.Token)
	if err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	err = h.ctrl.CheckForgotPasswordEmail(r.Context(), req.Password, uidb64, intToken)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to send email", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) sendForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.sendForgotPasswordEmail.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	req := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	err := h.ctrl.SendForgotPasswordEmail(r.Context(), req.Email)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to send email", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.updateMe.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(r.Context().Value("uid").(string))
	if err != nil {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, err)
		return
	}

	req := &model.User{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	if err = validation.UserValidation(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.UpdateUser(r.Context(), uid, req)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to update user", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, res)
}

func (h *Handler) checkEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkEmail.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	u := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	if u.Email == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, handler.ErrMissingEmail)
		return
	}

	isExist, err := h.ctrl.IsUserExist(r.Context(), u.Email)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to check existence of email", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, struct {
		IsExist bool `json:"is_exist"`
	}{
		IsExist: isExist,
	})
}

type loginCodeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) sendLoginCode(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.sendLoginCode.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	data := &loginCodeRequest{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	email, pass := data.Email, data.Password
	if err := validation.ValidateEmail(email); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	if email == "" || pass == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, handler.ErrEmailAndPasswordRequired)
		return
	}

	err = h.ctrl.SendLoginCode(r.Context(), email, pass)
	if err != nil && err == controller.ErrInvalidCredentials {
		c = http.StatusNotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to send login code", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "Login code sent successfully")
}

type checkLoginCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *Handler) checkLoginCode(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkLoginCode.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	data := &checkLoginCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to decode request", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	email, code := data.Email, data.Code
	if email == "" || code == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, handler.ErrEmailAndCodeRequired)
		return
	}

	if err := validation.ValidateEmail(email); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	loginCode, err := strconv.Atoi(code)
	if err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	access, refresh, err := h.ctrl.CheckLoginCode(r.Context(), email, loginCode)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to send login code", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    access,
		Expires:  time.Now().Add(auth.AccessTokenDuration),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    refresh,
		Expires:  time.Now().Add(auth.RefreshTokenDuration),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.me.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(r.Context().Value("uid").(string))
	if err != nil {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, err)
		return
	}

	u, err := h.ctrl.GetUserByID(r.Context(), uid)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug("failed to get user", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.logout.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	utils.SuccessResponse(w, c, "OK")
}
