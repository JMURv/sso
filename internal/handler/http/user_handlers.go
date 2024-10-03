package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	controller "github.com/JMURv/sso/internal/controller"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

func RegisterUserRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/api/users/search", h.userSearch).Methods(http.MethodGet)
	r.HandleFunc("/api/users", h.listUsers).Methods(http.MethodGet)
	r.HandleFunc("/api/users", h.register).Methods(http.MethodPost)
	r.HandleFunc("/api/users/{id}", h.getUser).Methods(http.MethodGet)
	r.HandleFunc("/api/users/{id}", middlewareFunc(h.updateUser, h.authMiddleware)).Methods(http.MethodPut)
	r.HandleFunc("/api/users/{id}", middlewareFunc(h.deleteUser, h.authMiddleware)).Methods(http.MethodDelete)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusCreated
	const op = "sso.register.handler"
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
			utils.ErrResponse(w, c, err)
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

	utils.SuccessResponse(w, c, uid)
}

func (h *Handler) userSearch(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.search.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		size = 10
	}

	res, err := h.ctrl.UserSearch(r.Context(), query, page, size)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessPaginatedResponse(w, c, res)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.listUsers.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		size = consts.DefaultPageSize
	}

	res, err := h.ctrl.ListUsers(r.Context(), page, size)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessPaginatedResponse(w, c, res)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.getUser.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(mux.Vars(r)["id"])
	if uid == uuid.Nil || err != nil {
		c = http.StatusBadRequest
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
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, res)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.updateUser.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(mux.Vars(r)["id"])
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
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusNoContent
	const op = "sso.deleteUser.handler"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := uuid.Parse(mux.Vars(r)["id"])
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
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, "OK")
}
