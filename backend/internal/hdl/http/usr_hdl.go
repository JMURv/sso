package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	_ "github.com/JMURv/sso/internal/models"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func (h *Handler) RegisterUserRoutes() {
	h.router.Post("/users/exists", h.existsUser)
	h.router.With(mid.Auth(h.au)).Get("/users/me", h.getMe)
	h.router.With(mid.Auth(h.au)).Put("/users/me", h.updateMe)
	h.router.Get("/users", h.listUsers)
	h.router.Post("/users", h.createUser)
	h.router.Get("/users/{id}", h.getUser)
	h.router.With(mid.Auth(h.au)).Put("/users/{id}", h.updateUser)
	h.router.With(mid.Auth(h.au)).Delete("/users/{id}", h.deleteUser)
}

// existsUser godoc
//
//	@Summary		Check if a user exists by email
//	@Description	Returns 200 if user exists, 404 otherwise
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CheckEmailRequest	true	"Email payload"
//	@Success		200		{object}	dto.ExistsUserResponse
//	@Failure		404		{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/exists [post]
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

// listUsers godoc
//
//	@Summary		List all users
//	@Description	Retrieve a paginated list of users with optional filters
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Page number"	default(1)
//	@Param			size	query		int	false	"Page size"		default(20)
//	@Success		200		{array}		dto.PaginatedUserResponse
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users [get]
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

// getMe godoc
//
//	@Summary		Retrieve current user profile
//	@Description	Returns the authenticated user's profile
//	@Tags			User
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer token, e.g. 'Bearer {jwt}'"
//	@Success		200				{object}	models.User
//	@Failure		401				{object}	utils.ErrorsResponse	"unauthorized"
//	@Failure		404				{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/me [get]
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Error(
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

// updateMe godoc
//
//	@Summary		Update current user
//	@Description	Updates user profile and avatar
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string					true	"User UUID"
//	@Param			data			formData	string					true	"JSON payload in 'data' field"
//	@Param			avatar			formData	file					false	"Avatar image file"
//	@Param			Authorization	header		string					true	"Bearer token, e.g. 'Bearer {jwt}'"
//	@Success		200				{object}	nil						"OK"
//	@Failure		400				{object}	utils.ErrorsResponse	"bad request"
//	@Failure		401				{object}	utils.ErrorsResponse	"unauthorized"
//	@Failure		404				{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/me [put]
func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	var err error
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if uid == uuid.Nil || !ok {
		zap.L().Error(
			hdl.ErrFailedToParseUUID.Error(),
			zap.Any("uid", r.Context().Value("uid")),
		)
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrFailedToParseUUID)
		return
	}

	if err = r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		zap.L().Info("failed to parse multipart form", zap.Error(err))
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

	err = h.ctrl.UpdateMe(r.Context(), uid, req, fileReq)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		utils.ErrResponse(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

// getUser godoc
//
//	@Summary		Get user by ID
//	@Description	Retrieve a user by their UUID
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"User UUID"
//	@Success		200	{object}	models.User
//	@Failure		400	{object}	utils.ErrorsResponse	"invalid UUID"
//	@Failure		404	{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500	{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/{id} [get]
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if uid == uuid.Nil || err != nil {
		zap.L().Error(
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

// createUser godoc
//
//	@Summary		Create a new user
//	@Description	Creates a user with optional avatar upload
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			data	formData	string	true	"JSON payload in 'data' field"
//	@Param			avatar	formData	file	false	"Avatar image file"
//	@Success		201		{object}	dto.CreateUserResponse
//	@Failure		400		{object}	utils.ErrorsResponse	"bad request or file too large"
//	@Failure		409		{object}	utils.ErrorsResponse	"user already exists"
//	@Failure		500		{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users [post]
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

// updateUser godoc
//
//	@Summary		Update an existing user
//	@Description	Updates user profile and avatar
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string					true	"User UUID"
//	@Param			data			formData	string					true	"JSON payload in 'data' field"
//	@Param			avatar			formData	file					false	"Avatar image file"
//	@Param			Authorization	header		string					true	"Bearer token, e.g. 'Bearer {jwt}'"
//	@Success		200				{object}	nil						"OK"
//	@Failure		400				{object}	utils.ErrorsResponse	"bad request"
//	@Failure		401				{object}	utils.ErrorsResponse	"unauthorized"
//	@Failure		404				{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/{id} [put]
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || uid == uuid.Nil {
		zap.L().Error(
			hdl.ErrFailedToParseUUID.Error(),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrFailedToParseUUID)
		return
	}

	if err = r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		zap.L().Info("failed to parse multipart form", zap.Error(err))
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

// deleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Removes a user by UUID
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string					true	"User UUID"
//	@Param			Authorization	header		string					true	"Bearer token, e.g. 'Bearer {jwt}'"
//	@Success		204				{object}	nil						"No Content"
//	@Failure		401				{object}	utils.ErrorsResponse	"unauthorized"
//	@Failure		404				{object}	utils.ErrorsResponse	"user not found"
//	@Failure		500				{object}	utils.ErrorsResponse	"internal error"
//	@Router			/users/{id} [delete]
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		zap.L().Error(
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
