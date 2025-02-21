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
	md "github.com/JMURv/sso/internal/models"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
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
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/sso/me", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mid.Apply(h.me, mid.Auth)
			case http.MethodPut:
				mid.Apply(h.updateMe, mid.Auth)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc("/api/sso/auth", h.authenticate)
	mux.HandleFunc("/api/sso/send-login-code", h.sendLoginCode)
	mux.HandleFunc("/api/sso/check-login-code", h.checkLoginCode)
	mux.HandleFunc("/api/sso/check-email", h.checkEmail)
	mux.HandleFunc("/api/sso/logout", mid.Apply(h.logout, mid.Auth))
	mux.HandleFunc("/api/sso/support", mid.Apply(h.sendSupportEmail, mid.Auth))
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
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	req := &dto.EmailAndPasswordRequest{}
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

	if err := validation.LoginAndPasswordRequest(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	res, err := h.ctrl.Authenticate(ctx, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
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

func (h *Handler) parseClaims(w http.ResponseWriter, r *http.Request) {
	const op = "sso.parseClaims.hdl"
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
			zap.Any("req", req),
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

	res, err := h.ctrl.ParseClaims(ctx, req)
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

func (h *Handler) getUserByToken(w http.ResponseWriter, r *http.Request) {
	const op = "sso.getUserByToken.hdl"
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
			zap.Any("req", req),
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

	res, err := h.ctrl.GetUserByToken(ctx, req)
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

func (h *Handler) sendSupportEmail(w http.ResponseWriter, r *http.Request) {
	const op = "sso.sendSupportEmail.hdl"
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

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToGetUUID.Error(),
			zap.String("op", op), zap.Any("uid", ctx.Value("uid")),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	req := &dto.SendSupportEmailRequest{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
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

	err = h.ctrl.SendSupportEmail(ctx, uid, req)
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
	const op = "sso.checkForgotPasswordEmail.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPut {
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
			zap.Any("req", req),
			zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrDecodeRequest)
		return
	}

	err := h.ctrl.CheckForgotPasswordEmail(ctx, req)
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

func (h *Handler) sendForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	const op = "sso.sendForgotPasswordEmail.hdl"
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

	req := &md.User{}
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

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	const op = "sso.updateMe.hdl"
	s, c := time.Now(), http.StatusOK
	span, ctx := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPut {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
		return
	}

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToGetUUID.Error(),
			zap.String("op", op), zap.Any("uid", ctx.Value("uid")),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	req := &md.User{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
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

	if err = validation.UserValidation(req); err != nil {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, err)
		return
	}

	err = h.ctrl.UpdateUser(ctx, uid, req)
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

func (h *Handler) checkEmail(w http.ResponseWriter, r *http.Request) {
	const op = "sso.checkEmail.hdl"
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

	req := &dto.CheckEmailRequest{}
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

	if req.Email == "" {
		c = http.StatusBadRequest
		utils.ErrResponse(w, c, validation.ErrMissingEmail)
		return
	}

	res, err := h.ctrl.IsUserExist(ctx, req.Email)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
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

func (h *Handler) sendLoginCode(w http.ResponseWriter, r *http.Request) {
	const op = "sso.sendLoginCode.hdl"
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
	const op = "sso.checkLoginCode.hdl"
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

	access, refresh, err := h.ctrl.CheckLoginCode(ctx, req.Email, req.Code)
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

	utils.StatusResponse(w, c)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	const op = "sso.me.hdl"
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

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToGetUUID.Error(),
			zap.String("op", op), zap.Any("uid", ctx.Value("uid")),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c = http.StatusInternalServerError
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	u, err := h.ctrl.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = http.StatusNotFound
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, c, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	const op = "sso.logout.hdl"
	s, c := time.Now(), http.StatusOK
	span, _ := opentracing.StartSpanFromContext(r.Context(), op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodPost {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, ErrMethodNotAllowed)
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
