package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handler) RegisterDeviceRoutes() {
	h.router.With(mid.Auth(h.au)).Get("/api/device", h.listDevices)
	h.router.With(mid.Auth(h.au)).Get("/api/device/{id}", h.getDevice)
	h.router.With(mid.Auth(h.au), mid.CheckRights(h.ctrl)).Put("/api/device/{id}", h.updateDevice)
	h.router.With(mid.Auth(h.au), mid.CheckRights(h.ctrl)).Delete("/api/device/{id}", h.deleteDevice)
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
		return
	}

	res, err := h.ctrl.ListDevices(r.Context(), uid)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	dID := chi.URLParam(r, "id")
	if dID == "" {
		zap.L().Debug(
			hdl.ErrToRetrievePathArg.Error(),
			zap.String("path", r.URL.Path),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrToRetrievePathArg)
		return
	}

	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
		return
	}

	res, err := h.ctrl.GetDevice(r.Context(), uid, dID)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) updateDevice(w http.ResponseWriter, r *http.Request) {
	dID := chi.URLParam(r, "id")
	if dID == "" {
		zap.L().Debug(
			hdl.ErrToRetrievePathArg.Error(),
			zap.String("path", r.URL.Path),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrToRetrievePathArg)
		return
	}

	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
		return
	}

	req := &dto.UpdateDeviceRequest{}
	if ok = utils.ParseAndValidate(w, r, req); !ok {
		return
	}

	err := h.ctrl.UpdateDevice(r.Context(), uid, dID, req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	dID := chi.URLParam(r, "id")
	if dID == "" {
		zap.L().Debug(
			hdl.ErrToRetrievePathArg.Error(),
			zap.String("path", r.URL.Path),
		)
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrToRetrievePathArg)
		return
	}

	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
		return
	}

	err := h.ctrl.DeleteDevice(r.Context(), uid, dID)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusNoContent)
}
