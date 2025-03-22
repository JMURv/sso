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
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

func RegisterPermRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc(
		"/api/perm", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.listPerms(w, r)
			case http.MethodPost:
				h.createPerm(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/perm/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.getPerm(w, r)
			case http.MethodPut:
				mid.Apply(h.updatePerm, mid.Auth(au))(w, r)
			case http.MethodDelete:
				mid.Apply(h.deletePerm, mid.Auth(au))(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)
}

func (h *Handler) listPerms(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = config.DefaultPage
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = config.DefaultSize
	}

	res, err := h.ctrl.ListPermissions(r.Context(), page, size)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) createPerm(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreatePermissionRequest{}
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

	res, err := h.ctrl.CreatePerm(r.Context(), req)
	if err != nil && errors.Is(err, ctrl.ErrAlreadyExists) {
		utils.ErrResponse(w, http.StatusConflict, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) getPerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perm/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	res, err := h.ctrl.GetPermission(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) updatePerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perm/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	req := &dto.UpdatePermissionRequest{}
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

	err = h.ctrl.UpdatePerm(r.Context(), uid, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) deletePerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perm/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	err = h.ctrl.DeletePerm(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusNoContent)
}
