package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *Handler) RegisterRoleRoutes() {
	h.router.Get("/roles", h.listRoles)
	h.router.With(mid.Auth(h.au)).Post("/roles", h.createRole)

	h.router.Get("/roles/{id}", h.getRole)
	h.router.With(mid.Auth(h.au)).Put("/roles/{id}", h.updateRole)
	h.router.With(mid.Auth(h.au)).Delete("/roles/{id}", h.deleteRole)
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	filters := utils.ParseFiltersByURL(r)
	page, size := utils.ParsePaginationValues(r)
	res, err := h.ctrl.ListRoles(r.Context(), page, size, filters)
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
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
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
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
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
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
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
