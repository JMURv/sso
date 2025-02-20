package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/handler"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func RegisterAuthRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/sso/parse", h.parseClaims)
	mux.HandleFunc("/api/sso/user", h.getUserByToken)

	mux.HandleFunc(
		"/api/sso/recovery", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				h.sendForgotPasswordEmail(w, r)
			case http.MethodPut:
				h.checkForgotPasswordEmail(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/sso/me", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				middlewareFunc(h.me, h.authMiddleware)
			case http.MethodPut:
				middlewareFunc(h.updateMe, h.authMiddleware)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc("/api/sso/auth", h.authenticate)
	mux.HandleFunc("/api/sso/send-login-code", h.sendLoginCode)
	mux.HandleFunc("/api/sso/check-login-code", h.checkLoginCode)
	mux.HandleFunc("/api/sso/check-email", h.checkEmail)
	mux.HandleFunc("/api/sso/logout", middlewareFunc(h.logout, h.authMiddleware))
	mux.HandleFunc("/api/sso/support", middlewareFunc(h.sendSupportEmail, h.authMiddleware))
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) {
	const op = "sso.authenticate.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	req := &dto.EmailAndPasswordRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if err := validation.LoginAndPasswordRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.Authenticate(ctx, req)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, res)
}

type TokenReq struct {
	Token string `json:"token"`
}

func (h *Handler) parseClaims(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.parseClaims.hdl"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	req := &TokenReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if req.Token == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	res, err := h.ctrl.ParseClaims(r.Context(), req.Token)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessResponse(w, c, res)

}

func (h *Handler) getUserByToken(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.getUserByToken.hdl"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	req := &TokenReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if req.Token == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	res, err := h.ctrl.GetUserByToken(r.Context(), req.Token)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessResponse(w, c, res)
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

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	uid, err := uuid.Parse(r.Context().Value("uid").(string))
	if err != nil {
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op), zap.Error(err),
		)
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	req := &sendSupportEmailRequest{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	err = h.ctrl.SendSupportEmail(r.Context(), uid, req.Theme, req.Text)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

type checkForgotPasswordEmailRequest struct {
	Password string    `json:"password"`
	Uidb64   uuid.UUID `json:"uidb64"`
	Token    int       `json:"token"`
}

func (h *Handler) checkForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkForgotPasswordEmail.handler"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPut {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	req := &checkForgotPasswordEmailRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	err := h.ctrl.CheckForgotPasswordEmail(r.Context(), req.Password, req.Uidb64, req.Token)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
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

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	req := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	err := h.ctrl.SendForgotPasswordEmail(r.Context(), req.Email)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
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

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPut {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	str, ok := r.Context().Value("uid").(string)
	if !ok {
		zap.L().Debug(
			"failed to get uid from context",
			zap.String("op", op),
		)
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, controller.ErrUnauthorized)
		return
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op), zap.Error(err),
		)
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	req := &model.User{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if err = validation.UserValidation(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	err = h.ctrl.UpdateUser(r.Context(), uid, req)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) checkEmail(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkEmail.handler"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	u := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if u.Email == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, handler.ErrMissingEmail)
		return
	}

	isExist, err := h.ctrl.IsUserExist(r.Context(), u.Email)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessPaginatedResponse(
		w, c, struct {
			IsExist bool `json:"is_exist"`
		}{
			IsExist: isExist,
		},
	)
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

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	data := &loginCodeRequest{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
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
	if err != nil && errors.Is(err, controller.ErrInvalidCredentials) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "Login code sent successfully")
}

type checkLoginCodeRequest struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

func (h *Handler) checkLoginCode(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.checkLoginCode.handler"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	data := &checkLoginCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	email, code := data.Email, data.Code
	if email == "" || code == 0 {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, handler.ErrEmailAndCodeRequired)
		return
	}

	if err := validation.ValidateEmail(email); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	access, refresh, err := h.ctrl.CheckLoginCode(r.Context(), email, code)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	http.SetCookie(
		w, &http.Cookie{
			Name:     "access",
			Value:    access,
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
			Value:    refresh,
			Expires:  time.Now().Add(auth.RefreshTokenDuration),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.me.handler"

	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodGet {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	str, ok := r.Context().Value("uid").(string)
	if !ok {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		c = http.StatusUnauthorized
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	u, err := h.ctrl.GetUserByID(r.Context(), uid)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
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

	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
			c = http.StatusInternalServerError
			utils.ErrResponse(w, c, controller.ErrInternalError)
		}
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
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

	utils.SuccessResponse(w, c, "OK")
}
