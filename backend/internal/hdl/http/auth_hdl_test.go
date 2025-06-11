package http

import (
	"bytes"
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/JMURv/sso/tests/mocks"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Authenticate(t *testing.T) {
	const uri = "/auth/jwt"
	mock := gomock.NewController(t)
	defer mock.Finish()

	testErr := errors.New("test-err")
	mctrl := mocks.NewMockAppCtrl(mock)
	auth := mocks.NewMockCore(mock)
	h := New(mctrl, auth)

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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ErrMethodNotAllowed.Error(), res.Errors[0])
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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Errors[0])
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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, "required rule", res.Errors[0])
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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, "required rule", res.Errors[0])
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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrNotFound.Error(), res.Errors[0])
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.DeviceRequest{
						IP: "",
						UA: "",
					}, &dto.EmailAndPasswordRequest{
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
				res := &utils.ErrorsResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, testErr.Error(), res.Errors[0])
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.DeviceRequest{
						IP: "",
						UA: "",
					}, &dto.EmailAndPasswordRequest{
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
				res := &dto.TokenPair{Access: "token", Refresh: "token"}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, "token", res.Access)
				assert.Equal(t, "token", res.Refresh)
			},
			expect: func() {
				mctrl.EXPECT().Authenticate(
					gomock.Any(), &dto.DeviceRequest{
						IP: "",
						UA: "",
					}, &dto.EmailAndPasswordRequest{
						Email:    "email",
						Password: "password",
					},
				).Return(&dto.TokenPair{Access: "token", Refresh: "token"}, nil)
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
