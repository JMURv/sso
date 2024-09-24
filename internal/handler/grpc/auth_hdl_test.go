package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	m2 "github.com/JMURv/sso/internal/controller/mocks"
	"github.com/JMURv/sso/internal/handler/grpc/mocks"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestSendLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Case 1: Request is nil or missing fields
	t.Run("Invalid request", func(t *testing.T) {
		invalidReq := &pb.SendLoginCodeReq{Email: "", Password: ""}
		res, err := h.SendLoginCode(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Invalid email format
	t.Run("Invalid email", func(t *testing.T) {
		invalidReq := &pb.SendLoginCodeReq{Email: "invalid-email", Password: "test123"}
		res, err := h.SendLoginCode(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Invalid credentials (Ctrl returns ErrInvalidCredentials)
	t.Run("Invalid credentials", func(t *testing.T) {
		req := &pb.SendLoginCodeReq{Email: "test@example.com", Password: "wrong-pass"}
		mockCtrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(ctrl.ErrInvalidCredentials)

		res, err := h.SendLoginCode(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 4: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.SendLoginCodeReq{Email: "test@example.com", Password: "test123"}
		mockCtrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(errors.New("internal error"))

		res, err := h.SendLoginCode(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 5: Success case
	t.Run("Success", func(t *testing.T) {
		req := &pb.SendLoginCodeReq{Email: "test@example.com", Password: "test123"}
		mockCtrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(nil)

		res, err := h.SendLoginCode(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
	})
}

func TestCheckLoginCode(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email or code)
	t.Run("Invalid request", func(t *testing.T) {
		invalidReq := &pb.CheckLoginCodeReq{Email: "", Code: 0}
		res, err := h.CheckLoginCode(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Invalid email format
	t.Run("Invalid email", func(t *testing.T) {
		invalidReq := &pb.CheckLoginCodeReq{Email: "invalid-email", Code: 1234}
		res, err := h.CheckLoginCode(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: User not found (Ctrl returns ErrNotFound)
	t.Run("User not found", func(t *testing.T) {
		req := &pb.CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
		mockCtrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return("", "", ctrl.ErrNotFound)

		res, err := h.CheckLoginCode(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 4: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
		mockCtrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return("", "", errors.New("internal error"))

		res, err := h.CheckLoginCode(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 5: Success case
	t.Run("Success", func(t *testing.T) {
		req := &pb.CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
		accessToken := "access-token"
		refreshToken := "refresh-token"
		mockCtrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return(accessToken, refreshToken, nil)

		res, err := h.CheckLoginCode(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
		assert.Equal(t, accessToken, res.Access)
		assert.Equal(t, refreshToken, res.Refresh)
	})
}

func TestCheckEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.EmailMsg{Email: ""}
		res, err := h.CheckEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: User not found
	t.Run("User not found", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, ctrl.ErrNotFound)

		res, err := h.CheckEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 3: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, errors.New("internal error"))

		res, err := h.CheckEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 4: Success - email exists
	t.Run("Success email exists", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(true, nil)

		res, err := h.CheckEmail(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
		assert.Equal(t, true, res.IsExist)
	})

	// Case 5: Success - email does not exist
	t.Run("Success email does not exist", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "nonexistent@example.com"}
		mockCtrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, nil)

		res, err := h.CheckEmail(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
		assert.Equal(t, false, res.IsExist)
	})
}

func TestLogout(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	// Case 1: Missing UID in context (Unauthenticated)
	t.Run("Missing UID in context", func(t *testing.T) {
		ctx := context.Background()
		res, err := h.Logout(ctx, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	// Case 2: Invalid UUID in context
	t.Run("Invalid UUID in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uid", "invalid-uuid")
		res, err := h.Logout(ctx, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Success case - valid UID
	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "uid", uuid.New().String())

		res, err := h.Logout(ctx, &pb.Empty{})

		assert.NotNil(t, res)
		assert.Nil(t, err)
	})
}

func TestSendForgotPasswordEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.EmailMsg{Email: ""}
		res, err := h.SendForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Invalid credentials
	t.Run("Invalid credentials", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(ctrl.ErrInvalidCredentials)

		res, err := h.SendForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(errors.New("internal error"))

		res, err := h.SendForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 4: Success - email sent
	t.Run("Success", func(t *testing.T) {
		req := &pb.EmailMsg{Email: "test@example.com"}
		mockCtrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(nil)

		res, err := h.SendForgotPasswordEmail(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
	})
}

func TestCheckForgotPasswordEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing password, uid, or token)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{Password: "", Uidb64: "", Token: ""}
		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Invalid UID
	t.Run("Invalid UID", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{Password: "password", Uidb64: "invalid-uuid", Token: "123456"}
		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Invalid token
	t.Run("Invalid token", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{Password: "password", Uidb64: uuid.New().String(), Token: "invalid-token"}
		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 4: Invalid code error from controller
	t.Run("Invalid code", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{
			Password: "password",
			Uidb64:   uuid.New().String(),
			Token:    "123456",
		}
		mockCtrl.EXPECT().CheckForgotPasswordEmail(gomock.Any(), req.Password, gomock.Any(), 123456).Return(ctrl.ErrCodeIsNotValid)

		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 5: User not found error from controller
	t.Run("User not found", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{
			Password: "password",
			Uidb64:   uuid.New().String(),
			Token:    "123456",
		}
		mockCtrl.EXPECT().CheckForgotPasswordEmail(gomock.Any(), req.Password, gomock.Any(), 123456).Return(ctrl.ErrNotFound)

		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 6: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{
			Password: "password",
			Uidb64:   uuid.New().String(),
			Token:    "123456",
		}
		mockCtrl.EXPECT().CheckForgotPasswordEmail(gomock.Any(), req.Password, gomock.Any(), 123456).Return(errors.New("internal error"))

		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 7: Success - password reset confirmed
	t.Run("Success", func(t *testing.T) {
		req := &pb.CheckForgotPasswordEmailReq{
			Password: "password",
			Uidb64:   uuid.New().String(),
			Token:    "123456",
		}
		mockCtrl.EXPECT().CheckForgotPasswordEmail(gomock.Any(), req.Password, gomock.Any(), 123456).Return(nil)

		res, err := h.CheckForgotPasswordEmail(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
	})
}

func TestSendSupportEmail(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()

	// Adding a valid UID to the context
	uid := uuid.New().String()
	ctx = context.WithValue(ctx, "uid", uid)

	// Case 1: Invalid request (missing theme or text)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.SendSupportEmailReq{Theme: "", Text: ""}
		res, err := h.SendSupportEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Invalid UID
	t.Run("Invalid UID", func(t *testing.T) {
		invalidCtx := context.WithValue(ctx, "uid", "invalid-uuid")
		req := &pb.SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
		res, err := h.SendSupportEmail(invalidCtx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: User not found error from controller
	t.Run("User not found", func(t *testing.T) {
		req := &pb.SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
		mockCtrl.EXPECT().SendSupportEmail(gomock.Any(), gomock.Any(), req.Theme, req.Text).Return(ctrl.ErrNotFound)

		res, err := h.SendSupportEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 4: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		req := &pb.SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
		mockCtrl.EXPECT().SendSupportEmail(gomock.Any(), gomock.Any(), req.Theme, req.Text).Return(errors.New("internal error"))

		res, err := h.SendSupportEmail(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 5: Success - email sent
	t.Run("Success", func(t *testing.T) {
		req := &pb.SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
		mockCtrl.EXPECT().SendSupportEmail(gomock.Any(), gomock.Any(), req.Theme, req.Text).Return(nil)

		res, err := h.SendSupportEmail(ctx, req)

		assert.NotNil(t, res)
		assert.Nil(t, err)
	})
}
func TestMe(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	// Create a valid context with a UID
	validUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "uid", validUID)

	// Case 1: Missing UID in context
	t.Run("Missing UID", func(t *testing.T) {
		ctxWithoutUID := context.Background()
		res, err := h.Me(ctxWithoutUID, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	// Case 2: Invalid UID
	t.Run("Invalid UID", func(t *testing.T) {
		invalidCtx := context.WithValue(context.Background(), "uid", "invalid-uuid")
		res, err := h.Me(invalidCtx, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: User not found
	t.Run("User not found", func(t *testing.T) {
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNotFound)

		res, err := h.Me(ctx, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 4: Internal error from controller
	t.Run("Internal error", func(t *testing.T) {
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))

		res, err := h.Me(ctx, &pb.Empty{})

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 5: Success - user retrieved
	t.Run("Success", func(t *testing.T) {
		expectedUser := &md.User{
			ID:    uuid.New(),
			Name:  "Test User",
			Email: "test@example.com",
		}
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(expectedUser, nil)

		res, err := h.Me(ctx, &pb.Empty{})

		assert.NotNil(t, res)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser.ID.String(), res.Id)
		assert.Equal(t, expectedUser.Name, res.Name)
		assert.Equal(t, expectedUser.Email, res.Email)
	})
}

func TestUpdateMe(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	// Create a valid context with a UID
	validUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "uid", validUID)

	// Case 1: Missing UID in context
	t.Run("Missing UID", func(t *testing.T) {
		ctxWithoutUID := context.Background()
		res, err := h.UpdateMe(ctxWithoutUID, &pb.User{})

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	// Case 2: Invalid UID
	t.Run("Invalid UID", func(t *testing.T) {
		invalidCtx := context.WithValue(context.Background(), "uid", "invalid-uuid")
		res, err := h.UpdateMe(invalidCtx, &pb.User{})

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Request is nil
	t.Run("Nil Request", func(t *testing.T) {
		res, err := h.UpdateMe(ctx, nil)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 4: Validation error
	t.Run("Validation Error", func(t *testing.T) {
		invalidUser := &pb.User{Name: "", Email: "invalid-email"}
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)

		res, err := h.UpdateMe(ctx, invalidUser)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 5: User not found
	t.Run("User Not Found", func(t *testing.T) {
		user := &pb.User{Name: "Test User", Email: "test@example.com"}
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNotFound)

		res, err := h.UpdateMe(ctx, user)

		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 6: Internal error from controller
	t.Run("Internal Error", func(t *testing.T) {
		user := &pb.User{Name: "Test User", Email: "test@example.com"}
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))

		res, err := h.UpdateMe(ctx, user)

		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	// Case 7: Success - user updated
	t.Run("Success", func(t *testing.T) {
		user := &pb.User{Name: "Test User", Email: "test@example.com"}
		updatedUser := &md.User{
			ID:    uuid.New(),
			Name:  "Updated User",
			Email: "updated@example.com",
		}
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(updatedUser, nil)

		res, err := h.UpdateMe(ctx, user)

		assert.NotNil(t, res)
		assert.Nil(t, err)
		assert.Equal(t, updatedUser.ID.String(), res.Id)
		assert.Equal(t, updatedUser.Name, res.Name)
		assert.Equal(t, updatedUser.Email, res.Email)
	})
}
