package http

import (
	"context"
	"errors"
	"github.com/JMURv/sso/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	authSvc := mocks.NewMockAuthService(mock)
	mctrl := mocks.NewMockCtrl(mock)
	hdl := New(authSvc, mctrl)

	go hdl.Start(8080)
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/api/health-check")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", resp.StatusCode)
	}
}

func TestClose(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	authSvc := mocks.NewMockAuthService(mock)
	mctrl := mocks.NewMockCtrl(mock)
	hdl := New(authSvc, mctrl)

	go hdl.Start(8080)
	time.Sleep(500 * time.Millisecond)

	if err := hdl.Close(); err != nil {
		t.Fatalf("Expected no error while closing, got %v", err)
	}

	resp, err := http.Get("http://localhost:8080/api/health-check")
	if err == nil {
		defer resp.Body.Close()
	}
}

func TestAuthMiddleware(t *testing.T) {
	const uri = "/api/sso/logout"
	mock := gomock.NewController(t)
	defer mock.Finish()

	authSvc := mocks.NewMockAuthService(mock)
	mctrl := mocks.NewMockCtrl(mock)
	h := New(authSvc, mctrl)

	t.Run(
		"No header", func(t *testing.T) {
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
		"Invalid token", func(t *testing.T) {
			ctx := context.Background()

			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "invalid")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			middlewareFunc(h.logout, h.authMiddleware)(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
		},
	)

	t.Run(
		"Err While VerifyToken", func(t *testing.T) {
			ctx := context.Background()

			authSvc.EXPECT().
				VerifyToken("invalid-token").
				Return(nil, errors.New("invalid token")).
				Times(1)

			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer invalid-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			middlewareFunc(h.logout, h.authMiddleware)(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
		},
	)
}

func TestTracingMiddleware(t *testing.T) {
	const uri = "/api/sso/logout"
	mock := gomock.NewController(t)
	defer mock.Finish()

	authSvc := mocks.NewMockAuthService(mock)
	mctrl := mocks.NewMockCtrl(mock)
	h := New(authSvc, mctrl)

	t.Run(
		"Success", func(t *testing.T) {
			ctx := context.Background()

			req := httptest.NewRequest(http.MethodPost, uri, nil)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			middlewareFunc(h.logout, h.tracingMiddleware)(w, req)
		},
	)

}
