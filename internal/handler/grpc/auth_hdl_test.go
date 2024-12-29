package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/mocks"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/grpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestHandler_GetUserByToken(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()
	expRes := &md.User{}

	t.Run(
		"Invalid request", func(t *testing.T) {
			invalidReq := &pb.SSO_StringMsg{String_: ""}
			res, err := h.GetUserByToken(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test").Return(expRes, nil).Times(1)
			res, err := h.GetUserByToken(ctx, &pb.SSO_StringMsg{String_: "test"})

			assert.Equal(t, utils.ModelToProto(expRes), res)
			assert.Equal(t, codes.OK, status.Code(err))
		},
	)

	t.Run(
		"Not Found", func(t *testing.T) {
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test").Return(nil, ctrl.ErrNotFound).Times(1)
			_, err := h.GetUserByToken(ctx, &pb.SSO_StringMsg{String_: "test"})

			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			newErr := errors.New("internal error")
			mctrl.EXPECT().GetUserByToken(gomock.Any(), "test").Return(nil, newErr).Times(1)
			_, err := h.GetUserByToken(ctx, &pb.SSO_StringMsg{String_: "test"})

			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestHandler_ValidateToken(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()
	expSuccess := &pb.SSO_ParseClaimsRes{
		Token: uuid.New().String(),
		Email: "test@example.com",
		Exp:   "test-exp",
	}

	t.Run(
		"Invalid request", func(t *testing.T) {
			invalidReq := &pb.SSO_StringMsg{String_: ""}
			res, err := h.ParseClaims(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().ParseClaims(gomock.Any(), "test").Return(
				map[string]any{
					"uid":   expSuccess.Token,
					"email": expSuccess.Email,
					"exp":   expSuccess.Exp,
				}, nil,
			).Times(1)
			res, err := h.ParseClaims(ctx, &pb.SSO_StringMsg{String_: "test"})

			assert.Equal(t, expSuccess, res)
			assert.Equal(t, codes.OK, status.Code(err))
		},
	)
}

func TestSendLoginCode(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	t.Run(
		"Invalid request", func(t *testing.T) {
			invalidReq := &pb.SSO_SendLoginCodeReq{Email: "", Password: ""}
			res, err := h.SendLoginCode(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Invalid email", func(t *testing.T) {
			invalidReq := &pb.SSO_SendLoginCodeReq{Email: "invalid-email", Password: "test123"}
			res, err := h.SendLoginCode(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Invalid credentials", func(t *testing.T) {
			req := &pb.SSO_SendLoginCodeReq{Email: "test@example.com", Password: "wrong-pass"}
			mctrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(ctrl.ErrInvalidCredentials)

			res, err := h.SendLoginCode(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			req := &pb.SSO_SendLoginCodeReq{Email: "test@example.com", Password: "test123"}
			mctrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(errors.New("internal error"))

			res, err := h.SendLoginCode(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_SendLoginCodeReq{Email: "test@example.com", Password: "test123"}
			mctrl.EXPECT().SendLoginCode(gomock.Any(), req.Email, req.Password).Return(nil)

			res, err := h.SendLoginCode(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}

func TestCheckLoginCode(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email or code)
	t.Run(
		"Invalid request", func(t *testing.T) {
			invalidReq := &pb.SSO_CheckLoginCodeReq{Email: "", Code: 0}
			res, err := h.CheckLoginCode(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: Invalid email format
	t.Run(
		"Invalid email", func(t *testing.T) {
			invalidReq := &pb.SSO_CheckLoginCodeReq{Email: "invalid-email", Code: 1234}
			res, err := h.CheckLoginCode(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 3: User not found (Ctrl returns ErrNotFound)
	t.Run(
		"User not found", func(t *testing.T) {
			req := &pb.SSO_CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
			mctrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return("", "", ctrl.ErrNotFound)

			res, err := h.CheckLoginCode(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	// Case 4: Internal error from controller
	t.Run(
		"Internal error", func(t *testing.T) {
			req := &pb.SSO_CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
			mctrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return(
				"",
				"",
				errors.New("internal error"),
			)

			res, err := h.CheckLoginCode(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	// Case 5: Success case
	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_CheckLoginCodeReq{Email: "test@example.com", Code: 1234}
			accessToken := "access-token"
			refreshToken := "refresh-token"
			mctrl.EXPECT().CheckLoginCode(gomock.Any(), req.Email, int(req.Code)).Return(
				accessToken,
				refreshToken,
				nil,
			)

			res, err := h.CheckLoginCode(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
			assert.Equal(t, accessToken, res.Access)
			assert.Equal(t, refreshToken, res.Refresh)
		},
	)
}

func TestCheckEmail(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email)
	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: ""}
			res, err := h.CheckEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: User not found
	t.Run(
		"User not found", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, ctrl.ErrNotFound)

			res, err := h.CheckEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	// Case 3: Internal error from controller
	t.Run(
		"Internal error", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, errors.New("internal error"))

			res, err := h.CheckEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	// Case 4: Success - email exists
	t.Run(
		"Success email exists", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(true, nil)

			res, err := h.CheckEmail(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
			assert.Equal(t, true, res.IsExist)
		},
	)

	// Case 5: Success - email does not exist
	t.Run(
		"Success email does not exist", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "nonexistent@example.com"}
			mctrl.EXPECT().IsUserExist(gomock.Any(), req.Email).Return(false, nil)

			res, err := h.CheckEmail(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
			assert.Equal(t, false, res.IsExist)
		},
	)
}

func TestLogout(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	// Case 1: Missing UID in context (Unauthenticated)
	t.Run(
		"Missing UID in context", func(t *testing.T) {
			ctx := context.Background()
			res, err := h.Logout(ctx, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	// Case 2: Invalid UUID in context
	t.Run(
		"Invalid UUID in context", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "uid", "invalid-uuid")
			res, err := h.Logout(ctx, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 3: Success case - valid UID
	t.Run(
		"Success", func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "uid", uuid.New().String())

			res, err := h.Logout(ctx, &pb.SSO_Empty{})

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}

func TestSendForgotPasswordEmail(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing email)
	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: ""}
			res, err := h.SendForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: Invalid credentials
	t.Run(
		"Invalid credentials", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(ctrl.ErrInvalidCredentials)

			res, err := h.SendForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 3: Internal error from controller
	t.Run(
		"Internal error", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(errors.New("internal error"))

			res, err := h.SendForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	// Case 4: Success - email sent
	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_EmailMsg{Email: "test@example.com"}
			mctrl.EXPECT().SendForgotPasswordEmail(gomock.Any(), req.Email).Return(nil)

			res, err := h.SendForgotPasswordEmail(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}

func TestCheckForgotPasswordEmail(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	// Case 1: Invalid request (missing password, uid, or token)
	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{Password: "", Uidb64: "", Token: ""}
			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: Invalid UID
	t.Run(
		"Invalid UID", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{Password: "password", Uidb64: "invalid-uuid", Token: "123456"}
			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 3: Invalid token
	t.Run(
		"Invalid token", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{
				Password: "password",
				Uidb64:   uuid.New().String(),
				Token:    "invalid-token",
			}
			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	// Case 4: Invalid code error from controller
	t.Run(
		"Invalid code", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{
				Password: "password",
				Uidb64:   uuid.New().String(),
				Token:    "123456",
			}
			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				req.Password,
				gomock.Any(),
				123456,
			).Return(ctrl.ErrCodeIsNotValid)

			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 5: User not found error from controller
	t.Run(
		"User not found", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{
				Password: "password",
				Uidb64:   uuid.New().String(),
				Token:    "123456",
			}
			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				req.Password,
				gomock.Any(),
				123456,
			).Return(ctrl.ErrNotFound)

			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	// Case 6: Internal error from controller
	t.Run(
		"Internal error", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{
				Password: "password",
				Uidb64:   uuid.New().String(),
				Token:    "123456",
			}
			mctrl.EXPECT().CheckForgotPasswordEmail(
				gomock.Any(),
				req.Password,
				gomock.Any(),
				123456,
			).Return(errors.New("internal error"))

			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	// Case 7: Success - password reset confirmed
	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_CheckForgotPasswordEmailReq{
				Password: "password",
				Uidb64:   uuid.New().String(),
				Token:    "123456",
			}
			mctrl.EXPECT().CheckForgotPasswordEmail(gomock.Any(), req.Password, gomock.Any(), 123456).Return(nil)

			res, err := h.CheckForgotPasswordEmail(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}

func TestSendSupportEmail(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	uid := uuid.New().String()
	ctx = context.WithValue(ctx, "uid", uid)

	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_SendSupportEmailReq{Theme: "", Text: ""}
			res, err := h.SendSupportEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Unauthorized", func(t *testing.T) {
			invalidCtx := context.Background()
			req := &pb.SSO_SendSupportEmailReq{Theme: "", Text: ""}
			res, err := h.SendSupportEmail(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid ID", func(t *testing.T) {
			invalidCtx := context.WithValue(ctx, "uid", "invalid-uuid")
			req := &pb.SSO_SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
			res, err := h.SendSupportEmail(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Not found", func(t *testing.T) {
			req := &pb.SSO_SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
			mctrl.EXPECT().SendSupportEmail(gomock.Any(), gomock.Any(), req.Theme, req.Text).Return(ctrl.ErrNotFound)

			res, err := h.SendSupportEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			req := &pb.SSO_SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
			mctrl.EXPECT().SendSupportEmail(
				gomock.Any(),
				gomock.Any(),
				req.Theme,
				req.Text,
			).Return(errors.New("internal error"))

			res, err := h.SendSupportEmail(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_SendSupportEmailReq{Theme: "Support Request", Text: "Need help with..."}
			mctrl.EXPECT().SendSupportEmail(gomock.Any(), gomock.Any(), req.Theme, req.Text).Return(nil)

			res, err := h.SendSupportEmail(ctx, req)

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}

func TestMe(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	validUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "uid", validUID)

	t.Run(
		"Missing ID", func(t *testing.T) {
			ctxWithoutUID := context.Background()
			res, err := h.Me(ctxWithoutUID, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid ID", func(t *testing.T) {
			invalidCtx := context.WithValue(context.Background(), "uid", "invalid-uuid")
			res, err := h.Me(invalidCtx, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Not found", func(t *testing.T) {
			mctrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNotFound)

			res, err := h.Me(ctx, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			mctrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))

			res, err := h.Me(ctx, &pb.SSO_Empty{})

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			expectedUser := &md.User{
				ID:    uuid.New(),
				Name:  "Test User",
				Email: "test@example.com",
			}
			mctrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(expectedUser, nil)

			res, err := h.Me(ctx, &pb.SSO_Empty{})

			assert.NotNil(t, res)
			assert.Nil(t, err)
			assert.Equal(t, expectedUser.ID.String(), res.Id)
			assert.Equal(t, expectedUser.Name, res.Name)
			assert.Equal(t, expectedUser.Email, res.Email)
		},
	)
}

func TestUpdateMe(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	validUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "uid", validUID)

	t.Run(
		"Missing UID", func(t *testing.T) {
			ctxWithoutUID := context.Background()
			res, err := h.UpdateMe(ctxWithoutUID, &pb.SSO_User{})

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid UID", func(t *testing.T) {
			invalidCtx := context.WithValue(context.Background(), "uid", "invalid-uuid")
			res, err := h.UpdateMe(invalidCtx, &pb.SSO_User{})

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Nil Request", func(t *testing.T) {
			res, err := h.UpdateMe(ctx, nil)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Validation Error", func(t *testing.T) {
			invalidUser := &pb.SSO_User{Name: "", Email: "invalid-email"}
			mctrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)

			res, err := h.UpdateMe(ctx, invalidUser)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Not Found", func(t *testing.T) {
			user := &pb.SSO_User{Name: "Test User", Email: "test@example.com"}
			mctrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(ctrl.ErrNotFound)

			res, err := h.UpdateMe(ctx, user)

			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			user := &pb.SSO_User{Name: "Test User", Email: "test@example.com"}
			mctrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

			res, err := h.UpdateMe(ctx, user)

			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			user := &pb.SSO_User{Name: "Test User", Email: "test@example.com"}
			mctrl.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			res, err := h.UpdateMe(ctx, user)

			assert.NotNil(t, res)
			assert.Nil(t, err)
		},
	)
}
