package http

import (
	"bytes"
	"context"
	"errors"
	ctrl "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/handler"
	"github.com/JMURv/sso/internal/validation"
	"github.com/JMURv/sso/mocks"
	"github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testErr = errors.New("test error")
var validEmail = "test@example.com"

func TestHandler_Authenticate(t *testing.T) {
	const uri = "/api/sso/auth"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrMethodNotAllowed",
			method: http.MethodPut,
			status: http.StatusMethodNotAllowed,
			payload: map[string]any{
				"email":    "email",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
			},
			expect: func() {},
		},
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email":    0,
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
			},
			expect: func() {},
		},
		{
			name:   "ErrMissingEmail",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email":    "",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, validation.ErrMissingEmail.Error(), res.Error)
			},
			expect: func() {},
		},
		{
			name:   "ErrMissingPass",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email":    "email",
				"password": "",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, validation.ErrMissingPass.Error(), res.Error)
			},
			expect: func() {},
		},
		{
			name:   "StatusNotFound",
			method: http.MethodPost,
			status: http.StatusNotFound,
			payload: map[string]any{
				"email":    "email",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.EmailAndPasswordRequest{
						Email:    "email",
						Password: "password",
					},
				).Return(nil, ctrl.ErrNotFound)
			},
		},
		{
			name:   "StatusInternalServerError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"email":    "email",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, testErr.Error(), res.Error)
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.EmailAndPasswordRequest{
						Email:    "email",
						Password: "password",
					},
				).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusOK,
			payload: map[string]any{
				"email":    "email",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.EmailAndPasswordResponse{Token: "token"}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, "token", res.Token)
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.EmailAndPasswordRequest{
						Email:    "email",
						Password: "password",
					},
				).Return(&dto.EmailAndPasswordResponse{Token: "token"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, err := json.Marshal(tt.payload)
				require.NoError(t, err)

				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.authenticate(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer func() {
					assert.Nil(t, w.Result().Body.Close())
				}()

				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_GetUserByToken(t *testing.T) {
	const uri = "/api/sso/get-user-by-token"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
			assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			b, err := json.Marshal(
				map[string]any{
					"token": 123,
				},
			)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest (empty token)", func(t *testing.T) {
			b, err := json.Marshal(&TokenReq{Token: ""})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test-token").Return(nil, ctrl.ErrNotFound)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test-token").Return(nil, testErr)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByToken(gomock.Any(), "test-token").
				DoAndReturn(
					func(ctx context.Context, token string) (*model.User, error) {
						panic("test panic")
					},
				)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test-token").Return(&model.User{}, nil)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.getUserByToken(w, req)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.NotNil(t, res.Data)
		},
	)
}

func TestHandler_ParseClaims(t *testing.T) {
	const uri = "/api/sso/parse"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	expSuccess := map[string]any{
		"uid":   uuid.New().String(),
		"email": "test@example.com",
		"exp":   "test-exp",
	}

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.parseClaims(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
			assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			b, err := json.Marshal(
				map[string]any{
					"token": 123,
				},
			)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.parseClaims(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest (empty token)", func(t *testing.T) {
			b, err := json.Marshal(&TokenReq{Token: ""})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.parseClaims(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			mctrl.EXPECT().
				ParseClaims(gomock.Any(), "test-token").
				DoAndReturn(
					func(ctx context.Context, token string) bool {
						panic("test panic")
					},
				)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.parseClaims(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().ParseClaims(gomock.Any(), "test-token").Return(
				expSuccess, nil,
			)
			b, err := json.Marshal(&TokenReq{Token: "test-token"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.parseClaims(w, req)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, expSuccess, res.Data)
		},
	)

}

func TestSendLoginCode(t *testing.T) {
	const uri = "/api/sso/send-login-code"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email":    1,
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Invalid email", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email":    "invalid-email",
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Empty pass", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email":    "example@mail.com",
					"password": "",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			assert.Equal(t, handler.ErrEmailAndPasswordRequired.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInvalidCredentials", func(t *testing.T) {
			mctrl.EXPECT().SendLoginCode(
				gomock.Any(),
				"test@example.com",
				"test-pass",
			).Return(ctrl.ErrInvalidCredentials)

			body, err := json.Marshal(
				map[string]any{
					"email":    "test@example.com",
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			mctrl.EXPECT().SendLoginCode(
				gomock.Any(),
				"test@example.com",
				"test-pass",
			).Return(errors.New("internal error"))

			body, err := json.Marshal(
				map[string]any{
					"email":    "test@example.com",
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			mctrl.EXPECT().SendLoginCode(
				gomock.Any(),
				"test@example.com",
				"test-pass",
			).DoAndReturn(
				func(ctx context.Context, email, password string) error {
					panic("panic")
				},
			)

			body, err := json.Marshal(
				map[string]any{
					"email":    "test@example.com",
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().SendLoginCode(
				gomock.Any(),
				"test@example.com",
				"test-pass",
			).Return(nil)

			body, err := json.Marshal(
				map[string]any{
					"email":    "test@example.com",
					"password": "test-pass",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendLoginCode(w, req)
			assert.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		},
	)
}

func TestCheckLoginCode(t *testing.T) {
	const uri = "/api/sso/check-login-code"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": 1,
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Invalid email", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": "invalid-email",
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Empty Code", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": "invalid-email",
					"code":  0,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, handler.ErrEmailAndCodeRequired.Error(), res.Error)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().
				CheckLoginCode(
					gomock.Any(), "test@example.com", 1234,
				).Return("", "", ctrl.ErrNotFound).Times(1)

			body, err := json.Marshal(
				map[string]any{
					"email": "test@example.com",
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			mctrl.EXPECT().
				CheckLoginCode(
					gomock.Any(), "test@example.com", 1234,
				).Return("", "", errors.New("internal error")).Times(1)

			body, err := json.Marshal(
				map[string]any{
					"email": "test@example.com",
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			mctrl.EXPECT().
				CheckLoginCode(
					gomock.Any(), "test@example.com", 1234,
				).DoAndReturn(
				func(ctx context.Context, email string, code int) (string, string, error) {
					panic("panic")
				},
			)

			body, err := json.Marshal(
				map[string]any{
					"email": "test@example.com",
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			accessToken := "access-token"
			refreshToken := "refresh-token"
			mctrl.EXPECT().
				CheckLoginCode(
					gomock.Any(), "test@example.com", 1234,
				).Return(
				accessToken,
				refreshToken,
				nil,
			)

			body, err := json.Marshal(
				map[string]any{
					"email": "test@example.com",
					"code":  1234,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkLoginCode(w, req)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		},
	)
}

func TestCheckEmail(t *testing.T) {
	const uri = "/api/sso/check-email"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": 123,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Empty Email", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": "",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"Not found", func(t *testing.T) {
			email := "test@example.com"
			mctrl.EXPECT().IsUserExist(gomock.Any(), email).Return(false, ctrl.ErrNotFound)
			body, err := json.Marshal(
				map[string]any{
					"email": email,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			email := "test@example.com"
			mctrl.EXPECT().IsUserExist(gomock.Any(), email).Return(false, errors.New("internal error"))
			body, err := json.Marshal(
				map[string]any{
					"email": email,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			email := "test@example.com"
			mctrl.EXPECT().IsUserExist(gomock.Any(), email).DoAndReturn(
				func(ctx context.Context, email string) (bool, error) {
					panic("test panic")
				},
			)
			body, err := json.Marshal(
				map[string]any{
					"email": email,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success email exists", func(t *testing.T) {
			email := "test@example.com"
			mctrl.EXPECT().IsUserExist(gomock.Any(), email).Return(true, nil)
			body, err := json.Marshal(
				map[string]any{
					"email": email,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			res := struct {
				IsExist bool `json:"is_exist"`
			}{}

			err = json.NewDecoder(w.Result().Body).Decode(&res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, true, res.IsExist)
		},
	)

	t.Run(
		"Success email does not exist", func(t *testing.T) {
			email := "test@example.com"
			mctrl.EXPECT().IsUserExist(gomock.Any(), email).Return(false, nil)
			body, err := json.Marshal(
				map[string]any{
					"email": email,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.checkEmail(w, req)

			res := struct {
				IsExist bool `json:"is_exist"`
			}{}

			err = json.NewDecoder(w.Result().Body).Decode(&res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, false, res.IsExist)
		},
	)
}

func TestLogout(t *testing.T) {
	const uri = "/api/sso/logout"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			ctx := context.Background()

			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.logout(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"Unauthorized", func(t *testing.T) {
			ctx := context.Background()

			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			middlewareFunc(h.logout, h.authMiddleware)(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			ctx := context.Background()
			auth.EXPECT().VerifyToken("token").Return(
				map[string]any{
					"uid": "test-uid",
				}, nil,
			)

			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			middlewareFunc(h.logout, h.authMiddleware)(w, req)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)

			cookies := w.Result().Cookies()
			require.Len(t, cookies, 2)
			assert.Equal(t, "access", cookies[0].Name)
			assert.Equal(t, "refresh", cookies[1].Name)
		},
	)
}

func TestSendForgotPasswordEmail(t *testing.T) {
	const uri = "/api/sso/recovery"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"email": 123,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), validEmail).Return(ctrl.ErrNotFound)
			body, err := json.Marshal(
				map[string]any{
					"email": validEmail,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), validEmail).Return(testErr)
			body, err := json.Marshal(
				map[string]any{
					"email": validEmail,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			b, err := json.Marshal(
				map[string]any{
					"email": validEmail,
				},
			)
			require.Nil(t, err)

			mctrl.EXPECT().
				SendForgotPasswordEmail(gomock.Any(), validEmail).
				DoAndReturn(
					func(context.Context, string) error {
						panic("something went wrong")
					},
				).
				Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {

			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), validEmail).Return(nil)
			body, err := json.Marshal(
				map[string]any{
					"email": validEmail,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendForgotPasswordEmail(w, req)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, "OK", res.Data)
		},
	)
}

func TestCheckForgotPasswordEmail(t *testing.T) {
	const uri = "/api/sso/recovery"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	uid := uuid.New()
	reqData := &checkForgotPasswordEmailRequest{
		Password: "123",
		Uidb64:   uid,
		Token:    123,
	}

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			body, err := json.Marshal(
				map[string]any{
					"password": 123,
					"uidb64":   uuid.New().String(),
					"token":    123,
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			body, err := json.Marshal(reqData)
			require.Nil(t, err)

			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				reqData.Password,
				reqData.Uidb64,
				reqData.Token,
			).Return(ctrl.ErrNotFound)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				reqData.Password,
				reqData.Uidb64,
				reqData.Token,
			).Return(testErr)

			body, err := json.Marshal(reqData)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			b, err := json.Marshal(reqData)
			require.Nil(t, err)

			mctrl.EXPECT().
				CheckForgotPasswordEmail(
					gomock.Any(),
					reqData.Password,
					reqData.Uidb64,
					reqData.Token,
				).
				DoAndReturn(
					func(context.Context, uuid.UUID, string, string) error {
						panic("something went wrong")
					},
				).
				Times(1)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				reqData.Password,
				reqData.Uidb64,
				reqData.Token,
			).Return(nil)

			body, err := json.Marshal(reqData)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.checkForgotPasswordEmail(w, req)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)

			assert.Equal(t, http.StatusOK, w.Result().StatusCode)
			assert.Equal(t, "OK", res.Data)
		},
	)
}

func TestSendSupportEmail(t *testing.T) {
	const uri = "/api/sso/support"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	uid := uuid.New()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "uid", uid.String())

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
		},
	)

	t.Run(
		"ErrParseUUID", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), "uid", "invalid"))

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			b, err := json.Marshal(
				map[string]any{
					"theme": 123,
					"text":  "test-text",
				},
			)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			data := &sendSupportEmailRequest{
				Theme: "",
				Text:  "",
			}

			mctrl.EXPECT().
				SendSupportEmail(gomock.Any(), uid, data.Theme, data.Text).
				Return(ctrl.ErrNotFound)

			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			data := &sendSupportEmailRequest{
				Theme: "",
				Text:  "",
			}

			mctrl.EXPECT().
				SendSupportEmail(gomock.Any(), uid, data.Theme, data.Text).
				Return(ctrl.ErrInternalError)

			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			data := &sendSupportEmailRequest{
				Theme: "",
				Text:  "",
			}

			b, err := json.Marshal(data)
			require.Nil(t, err)

			mctrl.EXPECT().
				SendSupportEmail(gomock.Any(), uid, data.Theme, data.Text).
				DoAndReturn(
					func(context.Context, uuid.UUID, string, string) error {
						panic("something went wrong")
					},
				).
				Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			data := &sendSupportEmailRequest{
				Theme: "",
				Text:  "",
			}

			mctrl.EXPECT().
				SendSupportEmail(gomock.Any(), uid, data.Theme, data.Text).
				Return(nil)

			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.sendSupportEmail(w, req)
			assert.Equal(t, http.StatusOK, w.Result().StatusCode)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, "OK", res.Data)
		},
	)
}

func TestMe(t *testing.T) {
	const uri = "/api/sso/me"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	validUID := uuid.New()
	ctx := context.WithValue(context.Background(), "uid", validUID.String())

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
		},
	)

	t.Run(
		"ErrParseUUID (empty context)", func(t *testing.T) {
			ctxWithoutUID := context.Background()
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctxWithoutUID)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrParseUUID (invalid uuid)", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(
				context.WithValue(context.Background(), "uid", validUID.String()+"invalid"),
			)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrNotFound", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), validUID).
				Return(nil, ctrl.ErrNotFound).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"ErrInternalError", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), validUID).
				Return(nil, ctrl.ErrInternalError).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), validUID).
				DoAndReturn(
					func(ctx context.Context, id uuid.UUID) (*model.User, error) {
						panic("something went wrong")
					},
				).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().
				GetUserByID(gomock.Any(), validUID).
				Return(&model.User{}, nil).
				Times(1)

			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.me(w, req)
			assert.Equal(t, http.StatusOK, w.Result().StatusCode)

			res := &utils.Response{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.NotNil(t, res.Data)
		},
	)

}

func TestUpdateMe(t *testing.T) {
	const uri = "/api/sso/me"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	validUID := uuid.New()
	ctx := context.WithValue(context.Background(), "uid", validUID.String())

	t.Run(
		"ErrMethodNotAllowed", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, handler.ErrMethodNotAllowed.Error(), res.Error)
		},
	)

	t.Run(
		"ErrParseUUID (empty context)", func(t *testing.T) {
			ctxWithoutUID := context.Background()
			req := httptest.NewRequest(http.MethodPut, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctxWithoutUID)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrUnauthorized.Error(), res.Error)
		},
	)

	t.Run(
		"ErrParseUUID (invalid uuid)", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(
				context.WithValue(context.Background(), "uid", validUID.String()+"invalid"),
			)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err := json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrParseUUID.Error(), res.Error)
		},
	)

	t.Run(
		"ErrDecodeRequest", func(t *testing.T) {
			b, err := json.Marshal(map[string]any{"name": 123})
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrDecodeRequest.Error(), res.Error)
		},
	)

	t.Run(
		"Validation Error", func(t *testing.T) {
			data := &model.User{Name: "test", Email: "invalid-email"}

			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, validation.ErrInvalidEmail.Error(), res.Error)
		},
	)

	t.Run(
		"Not Found", func(t *testing.T) {
			mctrl.EXPECT().
				UpdateUser(gomock.Any(), validUID, gomock.Any()).
				Return(ctrl.ErrNotFound)

			data := &model.User{Name: "test", Email: "test@example.com"}
			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrNotFound.Error(), res.Error)
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			mctrl.EXPECT().
				UpdateUser(gomock.Any(), validUID, gomock.Any()).
				Return(ctrl.ErrInternalError)

			data := &model.User{Name: "test", Email: "test@example.com"}
			b, err := json.Marshal(data)
			require.Nil(t, err)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Panic", func(t *testing.T) {
			data := &model.User{Name: "test", Email: "test@example.com"}
			b, err := json.Marshal(data)
			require.Nil(t, err)

			mctrl.EXPECT().
				UpdateUser(gomock.Any(), validUID, gomock.Any()).
				DoAndReturn(
					func(ctx context.Context, id uuid.UUID, u *model.User) error {
						panic("something went wrong")
					},
				).
				Times(1)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

			res := &utils.ErrorResponse{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, ctrl.ErrInternalError.Error(), res.Error)
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			data := &model.User{Name: "test", Email: "test@example.com"}
			b, err := json.Marshal(data)
			require.Nil(t, err)

			mctrl.EXPECT().
				UpdateUser(gomock.Any(), validUID, gomock.Any()).
				Return(nil)

			req := httptest.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h.updateMe(w, req)
			assert.Equal(t, http.StatusOK, w.Result().StatusCode)

			res := &utils.Response{}
			err = json.NewDecoder(w.Result().Body).Decode(res)
			require.Nil(t, err)
			assert.Equal(t, "OK", res.Data)
		},
	)
}
