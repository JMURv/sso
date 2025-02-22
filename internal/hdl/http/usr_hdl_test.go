package http

import (
	"bytes"
	"context"
	"errors"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/mocks"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_SearchUser(t *testing.T) {
	const uri = "/api/users/search"
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"Q < 3", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri+"?page=0&size=0&q=", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.searchUser(w, req)

			res := &model.PaginatedUser{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, 0, len(res.Data))
		},
	)

	t.Run(
		"MethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, uri+"?page=0&size=0&q=", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.searchUser(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			newErr := errors.New("internal error")
			mctrl.EXPECT().SearchUser(gomock.Any(), "testq", 1, consts.DefaultPageSize).Return(nil, newErr).Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+"?page=0&size=0&q=testq", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.searchUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, "internal error", res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().SearchUser(gomock.Any(), "testq", 1, 40).Return(&model.PaginatedUser{}, nil).Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+"?page=1&size=40&q=testq", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.searchUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			require.NotNil(t, res)
		},
	)

}

func TestHandler_ListUsers(t *testing.T) {
	const uri = "/api/users"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	t.Run(
		"Success", func(t *testing.T) {
			// Expect handler to wrap values to defaults (1, 40)
			mctrl.EXPECT().ListUsers(gomock.Any(), 1, 40).Return(&model.PaginatedUser{}, nil).Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+"?page=0&size=0", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.listUsers(w, req)

			res := &model.PaginatedUser{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, 0, len(res.Data))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			newErr := errors.New("internal error")
			mctrl.EXPECT().ListUsers(gomock.Any(), 1, 40).Return(nil, newErr).Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+"?page=1&size=40", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.listUsers(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, "internal error", res.Error)
		},
	)
}

func TestHandler_GetUser(t *testing.T) {
	const uri = "/api/users/"
	uid := uuid.New()
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrParseUUID", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri+uid.String()+"invalid", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.getUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), uid).
				Return(nil, controller.ErrNotFound).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.getUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			newErr := errors.New("internal error")
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), uid).
				Return(nil, newErr).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.getUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), uid).
				Return(&model.User{}, nil).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.getUser(w, req)

			res := &utils.Response{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.NotNil(t, res.Data)
		},
	)
}

func TestHandler_CreateUser(t *testing.T) {
	const uri = "/api/users"
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"Parse Multipart Form Error", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, uri, strings.NewReader("invalid-data"))
			req.Header.Set("Content-Type", "multipart/form-data")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.createUser(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Validation Error", func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("name", "testuser")
			_ = writer.WriteField("email", "invalid-email")
			_ = writer.WriteField("password", "password123")

			writer.Close()

			req := httptest.NewRequest(http.MethodPost, uri, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.createUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, validation.ErrInvalidEmail.Error(), res.Error)
		},
	)

	t.Run(
		"File Upload Success", func(t *testing.T) {
			uid := uuid.New()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("name", "testuser")
			_ = writer.WriteField("email", "user@example.com")
			_ = writer.WriteField("password", "password123")

			fw, err := writer.CreateFormFile("file", "test.png")
			assert.Nil(t, err)

			content := []byte("fake image data")
			_, err = fw.Write(content)
			assert.Nil(t, err)

			writer.Close()

			mctrl.EXPECT().CreateUser(gomock.Any(), gomock.Any(), "test.png", content).Return(
				uid,
				"accessToken",
				"refreshToken",
				nil,
			).Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.createUser(w, req)

			assert.Equal(t, http.StatusCreated, w.Result().StatusCode)

			cookies := w.Result().Cookies()
			require.Len(t, cookies, 2)
			assert.Equal(t, "access", cookies[0].Name)
			assert.Equal(t, "refresh", cookies[1].Name)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			require.NotNil(t, res.Data)
			assert.Equal(t, uid.String(), res.Data)
		},
	)

	t.Run(
		"Conflict Error", func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("name", "testuser")
			_ = writer.WriteField("email", "user@example.com")
			_ = writer.WriteField("password", "password123")

			writer.Close()

			mctrl.EXPECT().CreateUser(gomock.Any(), gomock.Any(), "", nil).Return(
				uuid.Nil,
				"",
				"",
				controller.ErrAlreadyExists,
			).Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.createUser(w, req)

			assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("name", "testuser")
			_ = writer.WriteField("email", "user@example.com")
			_ = writer.WriteField("password", "password123")

			writer.Close()

			mctrl.EXPECT().CreateUser(gomock.Any(), gomock.Any(), "", nil).Return(
				uuid.Nil,
				"",
				"",
				errors.New("internal error"),
			).Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.createUser(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		},
	)
}

func TestHandler_UpdateUser(t *testing.T) {
	const uri = "/api/users/"
	uid := uuid.New()
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrParseUUID", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, uri+uid.String()+"invalid", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"name":  "testuser-updated",
					"email": 123,
				},
			)
			req := httptest.NewRequest(http.MethodPut, uri+uid.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"Validation Error (invalid email)", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"name":  "testuser-updated",
					"email": "invalid@.com",
				},
			)
			req := httptest.NewRequest(http.MethodPut, uri+uid.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, validation.ErrInvalidEmail.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().UpdateUser(
				gomock.Any(), uid, &model.User{
					Name:  "testuser-updated",
					Email: "user@example.com",
				},
			).Return(controller.ErrNotFound).Times(1)

			body, err := json.Marshal(
				map[string]any{
					"name":  "testuser-updated",
					"email": "user@example.com",
				},
			)
			req := httptest.NewRequest(http.MethodPut, uri+uid.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			mctrl.EXPECT().UpdateUser(
				gomock.Any(), uid, &model.User{
					Name:  "testuser-updated",
					Email: "user@example.com",
				},
			).Return(controller.ErrInternalError).Times(1)

			body, err := json.Marshal(
				map[string]any{
					"name":  "testuser-updated",
					"email": "user@example.com",
				},
			)
			req := httptest.NewRequest(http.MethodPut, uri+uid.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().UpdateUser(
				gomock.Any(), uid, &model.User{
					Name:  "testuser-updated",
					Email: "user@example.com",
				},
			).Return(nil).Times(1)

			body, err := json.Marshal(
				map[string]any{
					"name":  "testuser-updated",
					"email": "user@example.com",
				},
			)
			req := httptest.NewRequest(http.MethodPut, uri+uid.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateUser(w, req)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, "OK", res.Data)
		},
	)
}

func TestHandler_DeleteUser(t *testing.T) {
	const uri = "/api/users/"
	uid := uuid.New()
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrParseUUID", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, uri+uid.String()+"invalid", nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.deleteUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().
				DeleteUser(gomock.Any(), uid).
				Return(controller.ErrNotFound).
				Times(1)

			req := httptest.NewRequest(http.MethodDelete, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.deleteUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			newErr := errors.New("internal error")
			mctrl.EXPECT().
				DeleteUser(gomock.Any(), uid).
				Return(newErr).
				Times(1)

			req := httptest.NewRequest(http.MethodDelete, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.deleteUser(w, req)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().
				DeleteUser(gomock.Any(), uid).
				Return(nil).
				Times(1)

			req := httptest.NewRequest(http.MethodDelete, uri+uid.String(), nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.deleteUser(w, req)

			res := &utils.Response{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
			require.NotNil(t, res)
			assert.Equal(t, "OK", res.Data)
		},
	)
}
