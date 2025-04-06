package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

// TODO: add search
func RegisterRoleRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc(
		"/api/roles/search", mid.Apply(
			h.searchRole,
			mid.AllowedMethods(http.MethodGet),
		),
	)

	mux.HandleFunc(
		"/api/roles", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.listRoles(w, r)
			case http.MethodPost:
				mid.Apply(h.createRole, mid.Auth(au))(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/roles/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.getRole(w, r)
			case http.MethodPut:
				mid.Apply(h.updateRole, mid.Auth(au))(w, r)
			case http.MethodDelete:
				mid.Apply(h.deleteRole, mid.Auth(au))(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)
}

func (h *Handler) searchRole(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		utils.SuccessResponse(w, http.StatusOK, dto.PaginatedRoleResponse{})
		return
	}

	page, size := utils.ParsePaginationValues(r)
	res, err := h.ctrl.SearchRole(r.Context(), query, page, size)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	page, size := utils.ParsePaginationValues(r)
	res, err := h.ctrl.ListRoles(r.Context(), page, size)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) createRole(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateRoleRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	res, err := h.ctrl.CreateRole(r.Context(), req)
	if err != nil && errors.Is(err, ctrl.ErrAlreadyExists) {
		utils.ErrResponse(w, http.StatusConflict, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) getRole(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/roles/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	res, err := h.ctrl.GetRole(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/roles/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	req := &dto.UpdateRoleRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	err = h.ctrl.UpdateRole(r.Context(), uid, req)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) deleteRole(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/roles/"), 10, 64)
	if err != nil {
		zap.L().Debug(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	err = h.ctrl.DeleteRole(r.Context(), uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusNoContent)
}
