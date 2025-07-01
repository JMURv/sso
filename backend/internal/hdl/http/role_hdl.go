package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	_ "github.com/JMURv/sso/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) RegisterRoleRoutes() {
	h.router.Get("/roles", h.listRoles)
	h.router.With(mid.Auth(h.au)).Post("/roles", h.createRole)

	h.router.Get("/roles/{id}", h.getRole)
	h.router.With(mid.Auth(h.au)).Put("/roles/{id}", h.updateRole)
	h.router.With(mid.Auth(h.au)).Delete("/roles/{id}", h.deleteRole)
}

// listRoles godoc
//
//	@Summary		List roles
//	@Description	Retrieve a paginated list of roles with optional filters
//	@Tags			Role
//	@Produce		json
//	@Param			page	query		int	false	"Page number"	default(1)
//	@Param			size	query		int	false	"Page size"		default(20)
//	@Success		200		{object}	dto.PaginatedRoleResponse
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/roles [get]
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

// createRole godoc
//
//	@Summary		Create a new role
//	@Description	Add a new role with associated permissions
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			body			body		dto.CreateRoleRequest	true	"Role details"
//	@Param			Authorization	header		string					true	"Authorization token"
//	@Success		201				{int}		"Role ID"
//	@Failure		400				{object}	utils.ErrorsResponse	"invalid request"
//	@Failure		409				{object}	utils.ErrorsResponse	"role already exists"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/roles [post]
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

// getRole godoc
//
//	@Summary		Get role by ID
//	@Description	Retrieve details of a specific role
//	@Tags			Role
//	@Produce		json
//	@Param			id	path		int	true	"Role ID"
//	@Success		200	{object}	models.Role
//	@Failure		400	{object}	utils.ErrorsResponse	"invalid ID"
//	@Failure		404	{object}	utils.ErrorsResponse	"role not found"
//	@Failure		500	{object}	utils.ErrorsResponse	"internal error"
//	@Router			/roles/{id} [get]
func (h *Handler) getRole(w http.ResponseWriter, r *http.Request) {
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

// updateRole godoc
//
//	@Summary		Update a role
//	@Description	Modify an existing role's name or permissions
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int						true	"Role ID"
//	@Param			body			body		dto.UpdateRoleRequest	true	"Updated role data"
//	@Param			Authorization	header		string					true	"Authorization token"
//	@Success		200				{object}	nil						"OK"
//	@Failure		400				{object}	utils.ErrorsResponse	"invalid request"
//	@Failure		404				{object}	utils.ErrorsResponse	"role not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/roles/{id} [put]
func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
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

// deleteRole godoc
//
//	@Summary		Delete a role
//	@Description	Remove a role by ID
//	@Tags			Role
//	@Param			id				path		int						true	"Role ID"
//	@Param			Authorization	header		string					true	"Authorization token"
//	@Success		204				{object}	nil						"No Content"
//	@Failure		400				{object}	utils.ErrorsResponse	"invalid ID"
//	@Failure		404				{object}	utils.ErrorsResponse	"role not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/roles/{id} [delete]
func (h *Handler) deleteRole(w http.ResponseWriter, r *http.Request) {
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
