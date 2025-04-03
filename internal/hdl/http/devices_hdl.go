package http

import (
	"errors"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// TODO: mid.Auth to mid.CheckOwner
func RegisterDeviceRoutes(mux *http.ServeMux, au auth.Core, h *Handler) {
	mux.HandleFunc(
		"/api/device", mid.Apply(
			h.listDevices,
			mid.AllowedMethods(http.MethodGet),
			mid.Auth(au),
		),
	)

	mux.HandleFunc(
		"/api/device/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mid.Apply(h.getDevice, mid.Auth(au))(w, r)
			case http.MethodPut:
				mid.Apply(h.updateDevice, mid.Auth(au))(w, r)
			case http.MethodDelete:
				mid.Apply(h.deleteDevice, mid.Auth(au))(w, r)
			default:
				utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			}
		},
	)
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
	dID := strings.TrimPrefix(r.URL.Path, "/api/device/")
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
	dID := strings.TrimPrefix(r.URL.Path, "/api/device/")
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
	dID := strings.TrimPrefix(r.URL.Path, "/api/device/")
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
