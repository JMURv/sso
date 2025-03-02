package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	ctrl "github.com/JMURv/sso/internal/ctrl"
	md "github.com/JMURv/sso/internal/models"
	gutils "github.com/JMURv/sso/internal/models/mapper"
	"github.com/JMURv/sso/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestUserSearch(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	page := uint64(1)
	size := uint64(10)
	query := "test"

	expectedData := &md.PaginatedUser{}

	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_SearchReq{Query: "", Page: 0, Size: 0}
			res, err := h.SearchUser(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_SearchReq{Query: query, Page: page, Size: size}

			mockCtrl.EXPECT().SearchUser(gomock.Any(), query, int(page), int(size)).Return(expectedData, nil).Times(1)

			res, err := h.SearchUser(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, int64(expectedData.TotalPages), res.TotalPages)
		},
	)

	t.Run(
		"Controller error", func(t *testing.T) {
			req := &pb.SSO_SearchReq{Query: query, Page: page, Size: size}

			mockCtrl.EXPECT().SearchUser(gomock.Any(), query, int(page), int(size)).Return(
				nil,
				errors.New("internal error"),
			).Times(1)

			res, err := h.SearchUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestListUsers(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	page := uint64(1)
	size := uint64(10)
	expectedData := &md.PaginatedUser{}

	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: 0, Size: 0}
			res, err := h.ListUsers(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: page, Size: size}

			mockCtrl.EXPECT().ListUsers(gomock.Any(), int(page), int(size)).Return(expectedData, nil).Times(1)

			res, err := h.ListUsers(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, expectedData.Count, res.Count)
		},
	)

	t.Run(
		"Controller error", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: page, Size: size}

			mockCtrl.EXPECT().ListUsers(gomock.Any(), int(page), int(size)).Return(
				nil,
				errors.New("internal error"),
			).Times(1)

			res, err := h.ListUsers(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestRegister(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	req := &pb.SSO_CreateUserReq{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "securepassword",
		File:     &pb.FileReq{Filename: "profile.jpg", File: []byte("filedata")},
	}
	protoUser := &md.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	idx := uuid.New()
	accessToken := "access-token"
	refreshToken := "refresh-token"

	// Case 1: Validation error
	t.Run(
		"Validation error", func(t *testing.T) {
			invalidReq := &pb.SSO_CreateUserReq{Name: "", Email: "", Password: ""}
			res, err := h.CreateUser(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: Success case - user created
	t.Run(
		"Success", func(t *testing.T) {
			mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
				Return(idx, accessToken, refreshToken, nil).Times(1)

			res, err := h.CreateUser(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, idx.String(), res.Uid)
			assert.Equal(t, accessToken, res.Access)
			assert.Equal(t, refreshToken, res.Refresh)
		},
	)

	// Case 3: User already exists
	t.Run(
		"User already exists", func(t *testing.T) {
			mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
				Return(uuid.Nil, "", "", ctrl.ErrAlreadyExists).Times(1)

			res, err := h.CreateUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
		},
	)

	// Case 4: Internal error
	t.Run(
		"Internal error", func(t *testing.T) {
			mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
				Return(uuid.Nil, "", "", errors.New("internal error")).Times(1)

			res, err := h.CreateUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestGetUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	validUUID := uuid.New().String()
	invalidUUID := "invalid-uuid"
	expectedUser := &md.User{}

	// Case 1: Invalid UUID
	t.Run(
		"Invalid UUID", func(t *testing.T) {
			req := &pb.SSO_UuidMsg{Uuid: invalidUUID}
			res, err := h.GetUser(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 2: Success case - user found
	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_UuidMsg{Uuid: validUUID}
			mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(expectedUser, nil).Times(1)

			res, err := h.GetUser(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
		},
	)

	// Case 3: User not found
	t.Run(
		"User Not Found", func(t *testing.T) {
			req := &pb.SSO_UuidMsg{Uuid: validUUID}
			mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNotFound).Times(1)

			res, err := h.GetUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	// Case 4: Internal error
	t.Run(
		"Internal Error", func(t *testing.T) {
			req := &pb.SSO_UuidMsg{Uuid: validUUID}
			mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error")).Times(1)

			res, err := h.GetUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestUpdateUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.SSO_UserWithUid{
		Uid:  uuid.New().String(),
		User: &pb.SSO_User{Name: "Test User", Email: "test@example.com"},
	}
	validUUID := uuid.MustParse(req.Uid)
	protoUser := gutils.ProtoToModel(req.User)

	t.Run(
		"Unauthorized", func(t *testing.T) {
			invalidCtx := context.Background() // No "uid" in the context
			res, err := h.UpdateUser(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid UUID", func(t *testing.T) {
			invalidReq := &pb.SSO_UserWithUid{Uid: "invalid-uuid"}
			res, err := h.UpdateUser(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Validation error", func(t *testing.T) {
			invalidUserReq := &pb.SSO_UserWithUid{
				Uid:  req.Uid,
				User: &pb.SSO_User{Name: ""}, // Invalid user data
			}
			res, err := h.UpdateUser(ctx, invalidUserReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"User updated", func(t *testing.T) {
			mockCtrl.EXPECT().UpdateUser(gomock.Any(), validUUID, protoUser).Return(nil).Times(1)

			res, err := h.UpdateUser(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, &pb.SSO_UuidMsg{Uuid: res.Uuid}, res)
		},
	)

	t.Run(
		"User not found", func(t *testing.T) {
			mockCtrl.EXPECT().UpdateUser(gomock.Any(), validUUID, protoUser).Return(ctrl.ErrNotFound).Times(1)

			res, err := h.UpdateUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			mockCtrl.EXPECT().UpdateUser(
				gomock.Any(),
				validUUID,
				protoUser,
			).Return(errors.New("internal error")).Times(1)

			res, err := h.UpdateUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestDeleteUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := mocks.NewMockAuthService(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.SSO_UuidMsg{Uuid: "123e4567-e89b-12d3-a456-426614174000"}
	validUUID := uuid.MustParse(req.Uuid)

	// Case 1: Unauthorized
	t.Run(
		"Unauthorized", func(t *testing.T) {
			invalidCtx := context.Background()
			res, err := h.DeleteUser(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	// Case 2: Invalid UUID
	t.Run(
		"Invalid UUID", func(t *testing.T) {
			invalidReq := &pb.SSO_UuidMsg{Uuid: "invalid-uuid"}
			res, err := h.DeleteUser(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	// Case 3: Success case - user deleted
	t.Run(
		"User deleted", func(t *testing.T) {
			mockCtrl.EXPECT().DeleteUser(gomock.Any(), validUUID).Return(nil).Times(1)

			res, err := h.DeleteUser(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
		},
	)

	// Case 4: Internal error
	t.Run(
		"Internal error", func(t *testing.T) {
			mockCtrl.EXPECT().DeleteUser(gomock.Any(), validUUID).Return(errors.New("internal error")).Times(1)

			res, err := h.DeleteUser(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}
