package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/handler"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterUserRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/users/search", h.searchUser)

	mux.HandleFunc(
		"/api/users", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.listUsers(w, r)
			case http.MethodPost:
				h.createUser(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/users/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.getUser(w, r)
			case http.MethodPut:
				middlewareFunc(h.updateUser, h.authMiddleware)
			case http.MethodDelete:
				middlewareFunc(h.deleteUser, h.authMiddleware)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusCreated
	const op = "sso.createUser.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to parse multipart form", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	u := &model.User{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := validation.NewUserValidation(u); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to validate user", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	var fileName string
	var bytes []byte
	file, handler, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		c = http.StatusBadRequest
		zap.L().Debug("failed to retrieve file", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	if file != nil {
		defer file.Close()
		bytes, err = io.ReadAll(file)
		if err != nil {
			c = http.StatusInternalServerError
			zap.L().Debug("failed to read file", zap.String("op", op), zap.Error(err))
			utils.ErrResponse(w, c, controller.ErrInternalError)
			return
		}
		fileName = handler.Filename
	}

	uid, access, refresh, err := h.ctrl.CreateUser(r.Context(), u, fileName, bytes)
	if err != nil && errors.Is(err, controller.ErrAlreadyExists) {
		c = http.StatusConflict
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

	utils.SuccessResponse(w, c, uid)
}

func (h *Handler) searchUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.search.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	if r.Method != http.MethodGet {
		c = http.StatusMethodNotAllowed
		utils.ErrResponse(w, c, handler.ErrMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		utils.SuccessPaginatedResponse(w, c, model.PaginatedUser{})
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = consts.DefaultPageSize
	}

	res, err := h.ctrl.SearchUser(r.Context(), query, page, size)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessPaginatedResponse(w, c, res)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.listUsers.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = consts.DefaultPageSize
	}

	res, err := h.ctrl.ListUsers(r.Context(), page, size)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, controller.ErrInternalError)
		return
	}

	utils.SuccessPaginatedResponse(w, c, res)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.getUser.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if uid == uuid.Nil || err != nil {
		c = http.StatusUnauthorized
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	res, err := h.ctrl.GetUserByID(r.Context(), uid)
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

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.updateUser.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if err != nil || uid == uuid.Nil {
		c = http.StatusUnauthorized
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
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

	if err = validation.UserValidation(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to validate obj",
			zap.String("op", op), zap.Error(err),
		)
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

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusNoContent
	const op = "sso.deleteUser.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if err != nil {
		c = http.StatusUnauthorized
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	err = h.ctrl.DeleteUser(r.Context(), uid)
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
