package http

import (
	"errors"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/handler"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/pkg/consts"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterPermRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc(
		"/api/perms", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.listPerms(w, r)
			case http.MethodPost:
				h.createPerm(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)

	mux.HandleFunc(
		"/api/perms/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				h.getPerm(w, r)
			case http.MethodPut:
				middlewareFunc(h.updatePerm, h.authMiddleware)
			case http.MethodDelete:
				middlewareFunc(h.deletePerm, h.authMiddleware)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, handler.ErrMethodNotAllowed)
			}
		},
	)
}

func (h *Handler) listPerms(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.listPerms.hdl"
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

	res, err := h.ctrl.ListPermissions(r.Context(), page, size)
	if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessPaginatedResponse(w, c, res)
}

func (h *Handler) createPerm(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusCreated
	const op = "sso.createPerm.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	req := &model.Permission{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if err := validation.PermValidation(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug("failed to validate user", zap.String("op", op), zap.Error(err))
		utils.ErrResponse(w, c, err)
		return
	}

	uid, err := h.ctrl.CreatePerm(r.Context(), req)
	if err != nil && errors.Is(err, controller.ErrAlreadyExists) {
		c = http.StatusConflict
		utils.ErrResponse(w, c, err)
		return
	} else if err != nil {
		c = http.StatusInternalServerError
		utils.ErrResponse(w, c, err)
		return
	}

	utils.SuccessResponse(w, c, uid)
}

func (h *Handler) getPerm(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.getPerm.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perms/"), 10, 64)
	if err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to parse id",
			zap.String("op", op),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	res, err := h.ctrl.GetPermission(r.Context(), uid)
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

func (h *Handler) updatePerm(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusOK
	const op = "sso.updatePerm.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perms/"), 10, 64)
	if err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to parse id",
			zap.String("op", op),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	req := &model.Permission{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to decode request",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, controller.ErrDecodeRequest)
		return
	}

	if err := validation.PermValidation(req); err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to validate obj",
			zap.String("op", op), zap.Error(err),
		)
		utils.ErrResponse(w, c, err)
		return
	}

	err = h.ctrl.UpdatePerm(r.Context(), uid, req)
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

func (h *Handler) deletePerm(w http.ResponseWriter, r *http.Request) {
	s, c := time.Now(), http.StatusNoContent
	const op = "sso.deletePerm.hdl"
	defer func() {
		metrics.ObserveRequest(time.Since(s), c, op)
	}()

	uid, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/perms/"), 10, 64)
	if err != nil {
		c = http.StatusBadRequest
		zap.L().Debug(
			"failed to parse id",
			zap.String("op", op),
		)
		utils.ErrResponse(w, c, controller.ErrParseUUID)
		return
	}

	err = h.ctrl.DeletePerm(r.Context(), uid)
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
