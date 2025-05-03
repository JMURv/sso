package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	_ "github.com/JMURv/sso/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *Handler) RegisterPermRoutes() {
	h.router.Get("/perm", h.listPerms)
	h.router.Post("/perm", h.createPerm)

	h.router.Get("/perm/{id}", h.getPerm)
	h.router.With(mid.Auth(h.au)).Put("/perm/{id}", h.updatePerm)
	h.router.With(mid.Auth(h.au)).Delete("/perm/{id}", h.deletePerm)
}

// listPerms godoc
//
//	@Summary		List permissions
//	@Description	Retrieve a paginated list of permissions with optional filters
//	@Tags			Permission
//	@Produce		json
//	@Param			page	query		int	false	"Page number"	default(1)
//	@Param			size	query		int	false	"Page size"		default(20)
//	@Success		200		{object}	dto.PaginatedPermissionResponse
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/perm [get]
func (h *Handler) listPerms(w http.ResponseWriter, r *http.Request) {
	filters := utils.ParseFiltersByURL(r)
	page, size := utils.ParsePaginationValues(r)
	res, err := h.ctrl.ListPermissions(r.Context(), page, size, filters)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

// createPerm godoc
//
//	@Summary		Create a new permission
//	@Description	Add a new permission to the system
//	@Tags			Permission
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreatePermissionRequest	true	"Permission details"
//	@Success		201		{int}		"Permission ID"
//	@Failure		400		{object}	utils.ErrorsResponse	"invalid request"
//	@Failure		409		{object}	utils.ErrorsResponse	"permission already exists"
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/perm [post]
func (h *Handler) createPerm(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreatePermissionRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
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

// getPerm godoc
//
//	@Summary		Get permission by ID
//	@Description	Retrieve a specific permission
//	@Tags			Permission
//	@Produce		json
//	@Param			id	path		int	true	"Permission ID"
//	@Success		200	{object}	models.Permission
//	@Failure		400	{object}	utils.ErrorsResponse	"invalid ID"
//	@Failure		404	{object}	utils.ErrorsResponse	"permission not found"
//	@Failure		500	{object}	utils.ErrorsResponse	"internal error"
//	@Router			/perm/{id} [get]
func (h *Handler) getPerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Error(
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

// updatePerm godoc
//
//	@Summary		Update a permission
//	@Description	Modify an existing permission
//	@Tags			Permission
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int							true	"Permission ID"
//	@Param			body			body		dto.UpdatePermissionRequest	true	"Updated permission data"
//	@Param			Authorization	header		string						true	"Authorization token"
//	@Success		200				{object}	nil							"OK"
//	@Failure		400				{object}	utils.ErrorsResponse		"invalid request"
//	@Failure		404				{object}	utils.ErrorsResponse		"permission not found"
//	@Failure		500				{object}	utils.ErrorsResponse		"internal error"
//	@Router			/perm/{id} [put]
func (h *Handler) updatePerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Error(
			ErrRetrievePathVars.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusBadRequest, ErrRetrievePathVars)
		return
	}

	req := &dto.UpdatePermissionRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
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

// deletePerm godoc

// @Summary		Delete a permission
// @Description	Remove a permission by ID
// @Tags			Permission
// @Param			id				path		int						true	"Permission ID"
// @Param			Authorization	header		string					true	"Authorization token"
// @Success		204				{object}	nil						"No Content"
// @Failure		400				{object}	utils.ErrorsResponse	"invalid ID"
// @Failure		404				{object}	utils.ErrorsResponse	"permission not found"
// @Failure		500				{object}	utils.ErrorsResponse	"internal error"
// @Router			/perm/{id} [delete]
func (h *Handler) deletePerm(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		zap.L().Error(
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
