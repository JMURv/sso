package http

//import (
//	"bytes"
//	"context"
//	"errors"
//	"fmt"
//	controller "github.com/JMURv/sso/internal/controller"
//	"github.com/JMURv/sso/internal/handler"
//	"github.com/JMURv/sso/internal/validation"
//	"github.com/JMURv/sso/mocks"
//	"github.com/JMURv/sso/pkg/model"
//	utils "github.com/JMURv/sso/pkg/utils/http"
//	"github.com/goccy/go-json"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"go.uber.org/mock/gomock"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"testing"
//)
//
//func TestHandler_ListPerm(t *testing.T) {
//	const uri = "/api/perm"
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	mctrl := mocks.NewMockCtrl(mock)
//	auth := mocks.NewMockAuthService(mock)
//	h := New(auth, mctrl)
//
//	ctx := context.Background()
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			// Expect handler to wrap values to defaults (1, 40)
//			mctrl.EXPECT().ListPermissions(gomock.Any(), 1, 40).Return(&model.PaginatedPermission{}, nil).Times(1)
//
//			req := httptest.NewRequest(http.MethodGet, uri+"?page=0&size=0", nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.listPerms(w, req)
//
//			res := &model.PaginatedPermission{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
//			assert.Equal(t, 0, len(res.Data))
//		},
//	)
//
//	t.Run(
//		"Internal Error", func(t *testing.T) {
//			newErr := errors.New("internal error")
//			mctrl.EXPECT().ListPermissions(gomock.Any(), 1, 40).Return(nil, newErr).Times(1)
//
//			req := httptest.NewRequest(http.MethodGet, uri+"?page=1&size=40", nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.listPerms(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, "internal error", res.Error)
//		},
//	)
//}
//
//func TestHandler_GetPerm(t *testing.T) {
//	const uri = "/api/perm/"
//	uid := uint64(1)
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mctrl := mocks.NewMockCtrl(mock)
//	auth := mocks.NewMockAuthService(mock)
//	h := New(auth, mctrl)
//
//	t.Run(
//		"ErrRetrievePathVars", func(t *testing.T) {
//			req := httptest.NewRequest(http.MethodGet, uri+fmt.Sprintf("%v", uid)+"invalid", nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.getPerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, handler.ErrRetrievePathVars.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrNotFound", func(t *testing.T) {
//			mctrl.EXPECT().
//				GetPermission(gomock.Any(), uid).
//				Return(nil, controller.ErrNotFound).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodGet, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.getPerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrInternalError", func(t *testing.T) {
//			newErr := errors.New("internal error")
//			mctrl.EXPECT().
//				GetPermission(gomock.Any(), uid).
//				Return(nil, newErr).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodGet, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.getPerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			mctrl.EXPECT().
//				GetPermission(gomock.Any(), uid).
//				Return(&model.Permission{}, nil).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodGet, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.getPerm(w, req)
//
//			res := &utils.Response{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.NotNil(t, res.Data)
//		},
//	)
//}
//
//func TestHandler_CreatePerm(t *testing.T) {
//	const uri = "/api/perm"
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	uid := uint64(1)
//	ctx := context.Background()
//	mctrl := mocks.NewMockCtrl(mock)
//	auth := mocks.NewMockAuthService(mock)
//	h := New(auth, mctrl)
//
//	t.Run(
//		"ErrDecodeRequest", func(t *testing.T) {
//			req := httptest.NewRequest(http.MethodPost, uri, strings.NewReader("invalid-data"))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.createPerm(w, req)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//		},
//	)
//
//	t.Run(
//		"Validation Error", func(t *testing.T) {
//			body, err := json.Marshal(
//				&model.Permission{
//					Name: "",
//				},
//			)
//			require.Nil(t, err)
//
//			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.createPerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, validation.ErrMissingName.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrAlreadyExists", func(t *testing.T) {
//			data := &model.Permission{
//				Name: "test-name",
//			}
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			mctrl.EXPECT().CreatePerm(gomock.Any(), data).Return(uint64(0), controller.ErrAlreadyExists).Times(1)
//
//			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.createPerm(w, req)
//
//			assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
//		},
//	)
//
//	t.Run(
//		"ErrInternal", func(t *testing.T) {
//			newErr := errors.New("internal error")
//			data := &model.Permission{
//				Name: "test-name",
//			}
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			mctrl.EXPECT().CreatePerm(gomock.Any(), data).Return(uint64(0), newErr).Times(1)
//
//			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.createPerm(w, req)
//
//			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			data := &model.Permission{
//				Name: "test-name",
//			}
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			mctrl.EXPECT().CreatePerm(gomock.Any(), data).Return(uid, nil).Times(1)
//
//			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.createPerm(w, req)
//
//			assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
//
//			res := &utils.Response{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			require.NotNil(t, res.Data)
//			assert.Equal(t, float64(uid), res.Data)
//		},
//	)
//}
//
//func TestHandler_UpdatePerm(t *testing.T) {
//	const uri = "/api/perm/"
//	uid := uint64(1)
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mctrl := mocks.NewMockCtrl(mock)
//	auth := mocks.NewMockAuthService(mock)
//	h := New(auth, mctrl)
//
//	t.Run(
//		"ErrRetrievePathVars", func(t *testing.T) {
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid)+"invalid", nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, handler.ErrRetrievePathVars.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrDecodeRequest", func(t *testing.T) {
//			body, err := json.Marshal(
//				map[string]any{
//					"name": 123,
//				},
//			)
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid), bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrDecodeRequest.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"Validation Error (missing name)", func(t *testing.T) {
//			body, err := json.Marshal(
//				map[string]any{
//					"name": "",
//				},
//			)
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid), bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, validation.ErrMissingName.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrNotFound", func(t *testing.T) {
//			data := &model.Permission{
//				Name: "name-updated",
//			}
//			mctrl.EXPECT().UpdatePerm(gomock.Any(), uid, data).Return(controller.ErrNotFound).Times(1)
//
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid), bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			require.Nil(t, err)
//
//			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrInternalError", func(t *testing.T) {
//			data := &model.Permission{
//				Name: "name-updated",
//			}
//			mctrl.EXPECT().UpdatePerm(gomock.Any(), uid, data).Return(controller.ErrInternalError).Times(1)
//
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid), bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			require.Nil(t, err)
//
//			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			data := &model.Permission{
//				Name: "name-updated",
//			}
//			mctrl.EXPECT().UpdatePerm(gomock.Any(), uid, data).Return(nil).Times(1)
//
//			body, err := json.Marshal(data)
//			require.Nil(t, err)
//
//			req := httptest.NewRequest(http.MethodPut, uri+fmt.Sprintf("%v", uid), bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.updatePerm(w, req)
//
//			res := &utils.Response{}
//			err = json.NewDecoder(w.Result().Body).Decode(res)
//			require.Nil(t, err)
//
//			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, "OK", res.Data)
//		},
//	)
//}
//
//func TestHandler_DeletePerm(t *testing.T) {
//	const uri = "/api/perm/"
//	uid := uint64(1)
//	mock := gomock.NewController(t)
//	defer mock.Finish()
//
//	ctx := context.Background()
//	mctrl := mocks.NewMockCtrl(mock)
//	auth := mocks.NewMockAuthService(mock)
//	h := New(auth, mctrl)
//
//	t.Run(
//		"ErrRetrievePathVars", func(t *testing.T) {
//			req := httptest.NewRequest(http.MethodDelete, uri+fmt.Sprintf("%v", uid)+"invalid", nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.deletePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, handler.ErrRetrievePathVars.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrNotFound", func(t *testing.T) {
//			mctrl.EXPECT().
//				DeletePerm(gomock.Any(), uid).
//				Return(controller.ErrNotFound).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodDelete, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.deletePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrNotFound.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"ErrInternalError", func(t *testing.T) {
//			newErr := errors.New("internal error")
//			mctrl.EXPECT().
//				DeletePerm(gomock.Any(), uid).
//				Return(newErr).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodDelete, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.deletePerm(w, req)
//
//			res := &utils.ErrorResponse{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, controller.ErrInternalError.Error(), res.Error)
//		},
//	)
//
//	t.Run(
//		"Success", func(t *testing.T) {
//			mctrl.EXPECT().
//				DeletePerm(gomock.Any(), uid).
//				Return(nil).
//				Times(1)
//
//			req := httptest.NewRequest(http.MethodDelete, uri+fmt.Sprintf("%v", uid), nil)
//			req.Header.Set("Content-Type", "application/json")
//			req = req.WithContext(ctx)
//
//			w := httptest.NewRecorder()
//			h.deletePerm(w, req)
//
//			res := &utils.Response{}
//			err := json.NewDecoder(w.Result().Body).Decode(res)
//			assert.Nil(t, err)
//
//			assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
//			require.NotNil(t, res)
//			assert.Equal(t, "OK", res.Data)
//		},
//	)
//}
