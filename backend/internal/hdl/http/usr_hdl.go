package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handler) RegisterUserRoutes() {
	h.router.Post("/api/users/exists", h.existsUser)
	h.router.With(mid.Auth(h.au)).Get("/api/users/me", h.getMe)

	h.router.Get("/api/users", h.listUsers)
	h.router.Post("/api/users", h.createUser)

	h.router.Get("/api/users/{id}", h.getUser)
	h.router.With(mid.Auth(h.au)).Put("/api/users/{id}", h.updateUser)
	h.router.With(mid.Auth(h.au)).Delete("/api/users/{id}", h.deleteUser)
}

func (h *Handler) existsUser(w http.ResponseWriter, r *http.Request) {
	req := &dto.CheckEmailRequest{}
	if ok := utils.ParseAndValidate(w, r, req); !ok {
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

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	page, size := utils.ParsePaginationValues(r)
	filters := utils.ParseFiltersByURL(r)

	res, err := h.ctrl.ListUsers(r.Context(), page, size, filters)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
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

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if uid == uuid.Nil || err != nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
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

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrFileTooLarge)
		return
	}

	req := &dto.CreateUserRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err := validator.New().Struct(req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	fileReq := &s3.UploadFileRequest{}
	if err := utils.ParseFileField(r, "avatar", fileReq); err != nil {
		if errors.Is(err, hdl.ErrInternal) {
			utils.ErrResponse(w, http.StatusInternalServerError, err)
			return
		}
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.CreateUser(r.Context(), req, fileReq)
	if err != nil {
		if errors.Is(err, ctrl.ErrAlreadyExists) {
			utils.ErrResponse(w, http.StatusConflict, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || uid == uuid.Nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrFailedToParseUUID)
		return
	}

	if err = r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		zap.L().Debug("failed to parse multipart form", zap.Error(err))
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	req := &dto.UpdateUserRequest{}
	if err = json.Unmarshal([]byte(r.FormValue("data")), req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return
	}

	if err = validator.New().Struct(req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	fileReq := &s3.UploadFileRequest{}
	if err = utils.ParseFileField(r, "avatar", fileReq); err != nil {
		if errors.Is(err, hdl.ErrInternal) {
			utils.ErrResponse(w, http.StatusInternalServerError, err)
			return
		}
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.ctrl.UpdateUser(r.Context(), uid, req, fileReq)
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
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		zap.L().Debug(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
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
