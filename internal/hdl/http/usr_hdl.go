package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/JMURv/sso/internal/hdl/validation"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

func RegisterUserRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc("/api/users/search", mid.Apply(
		h.searchUser,
		mid.AllowedMethods(http.MethodGet),
	))

	mux.HandleFunc("/api/users/exists", mid.Apply(
		h.existsUser,
		mid.AllowedMethods(http.MethodPost),
	))

	mux.HandleFunc(
		"/api/users", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.listUsers(w, r)
			case http.MethodPost:
				h.createUser(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/users/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.getUser(w, r)
			case http.MethodPut:
				mid.Apply(h.updateUser, mid.Auth(au))(w, r)
			case http.MethodDelete:
				mid.Apply(h.deleteUser, mid.Auth(au))(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)
}

func (h *Handler) existsUser(w http.ResponseWriter, r *http.Request) {
	req := &dto.CheckEmailRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.V.Struct(req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.IsUserExist(r.Context(), req.Email)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err := validation.V.Struct(req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.CreateUser(r.Context(), req)
	if err != nil && errors.Is(err, ctrl.ErrAlreadyExists) {
		utils.ErrResponse(w, http.StatusConflict, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) searchUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		utils.SuccessResponse(w, http.StatusOK, dto.PaginatedUserResponse{})
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = config.DefaultPage
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = config.DefaultSize
	}

	res, err := h.ctrl.SearchUser(r.Context(), query, page, size)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = config.DefaultPage
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = config.DefaultSize
	}

	res, err := h.ctrl.ListUsers(r.Context(), page, size)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if uid == uuid.Nil || err != nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrFailedToParseUUID)
		return
	}

	res, err := h.ctrl.GetUserByID(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if err != nil || uid == uuid.Nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrFailedToParseUUID)
		return
	}

	req := &dto.UpdateUserRequest{}
	if err = json.NewDecoder(r.Body).Decode(req); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err = validation.V.Struct(req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.ctrl.UpdateUser(r.Context(), uid, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/api/users/"))
	if err != nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrFailedToParseUUID)
		return
	}

	err = h.ctrl.DeleteUser(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusNoContent)
}
