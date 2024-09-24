package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	m2 "github.com/JMURv/sso/internal/controller/mocks"
	"github.com/JMURv/sso/internal/handler/grpc/mocks"
	md "github.com/JMURv/sso/pkg/model"
	gutils "github.com/JMURv/sso/pkg/utils/grpc"
	utils "github.com/JMURv/sso/pkg/utils/http"
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
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	page := uint64(1)
	size := uint64(10)
	query := "test"

	expectedData := &utils.PaginatedData{}

	// Case 1: Invalid request (query, page, or size is empty/zero)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.UserSearchReq{Query: "", Page: 0, Size: 0}
		res, err := h.UserSearch(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Controller returns user data successfully
	t.Run("Success", func(t *testing.T) {
		req := &pb.UserSearchReq{Query: query, Page: page, Size: size}

		mockCtrl.EXPECT().UserSearch(gomock.Any(), query, int(page), int(size)).Return(expectedData, nil).Times(1)

		res, err := h.UserSearch(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(expectedData.TotalPages), res.TotalPages)
	})

	// Case 3: Controller returns an internal error
	t.Run("Controller error", func(t *testing.T) {
		req := &pb.UserSearchReq{Query: query, Page: page, Size: size}

		mockCtrl.EXPECT().UserSearch(gomock.Any(), query, int(page), int(size)).Return(nil, errors.New("internal error")).Times(1)

		res, err := h.UserSearch(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestListUsers(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	page := uint64(1)
	size := uint64(10)
	expectedData := &utils.PaginatedData{}

	// Case 1: Invalid request (page or size = 0)
	t.Run("Invalid request", func(t *testing.T) {
		req := &pb.ListUsersReq{Page: 0, Size: 0}
		res, err := h.ListUsers(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Success case - controller returns user data
	t.Run("Success", func(t *testing.T) {
		req := &pb.ListUsersReq{Page: page, Size: size}

		mockCtrl.EXPECT().ListUsers(gomock.Any(), int(page), int(size)).Return(expectedData, nil).Times(1)

		res, err := h.ListUsers(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedData.Count, res.Count)
	})

	// Case 3: Controller error
	t.Run("Controller error", func(t *testing.T) {
		req := &pb.ListUsersReq{Page: page, Size: size}

		mockCtrl.EXPECT().ListUsers(gomock.Any(), int(page), int(size)).Return(nil, errors.New("internal error")).Times(1)

		res, err := h.ListUsers(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestRegister(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	req := &pb.RegisterReq{
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

	expectedUser := &md.User{}
	accessToken := "access-token"
	refreshToken := "refresh-token"

	// Case 1: Validation error
	t.Run("Validation error", func(t *testing.T) {
		invalidReq := &pb.RegisterReq{Name: "", Email: "", Password: ""}
		res, err := h.Register(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Success case - user created
	t.Run("Success", func(t *testing.T) {
		mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
			Return(expectedUser, accessToken, refreshToken, nil).Times(1)

		res, err := h.Register(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, accessToken, res.Access)
		assert.Equal(t, refreshToken, res.Refresh)
	})

	// Case 3: User already exists
	t.Run("User already exists", func(t *testing.T) {
		mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
			Return(nil, "", "", ctrl.ErrAlreadyExists).Times(1)

		res, err := h.Register(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	// Case 4: Internal error
	t.Run("Internal error", func(t *testing.T) {
		mockCtrl.EXPECT().CreateUser(gomock.Any(), protoUser, req.File.Filename, req.File.File).
			Return(nil, "", "", errors.New("internal error")).Times(1)

		res, err := h.Register(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestGetUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.Background()
	validUUID := uuid.New().String()
	invalidUUID := "invalid-uuid"
	expectedUser := &md.User{}

	// Case 1: Invalid UUID
	t.Run("Invalid UUID", func(t *testing.T) {
		req := &pb.UuidMsg{Uuid: invalidUUID}
		res, err := h.GetUser(ctx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 2: Success case - user found
	t.Run("Success", func(t *testing.T) {
		req := &pb.UuidMsg{Uuid: validUUID}
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(expectedUser, nil).Times(1)

		res, err := h.GetUser(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	// Case 3: User not found
	t.Run("User Not Found", func(t *testing.T) {
		req := &pb.UuidMsg{Uuid: validUUID}
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNotFound).Times(1)

		res, err := h.GetUser(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 4: Internal error
	t.Run("Internal Error", func(t *testing.T) {
		req := &pb.UuidMsg{Uuid: validUUID}
		mockCtrl.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error")).Times(1)

		res, err := h.GetUser(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestUpdateUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.UserWithUid{
		Uid:  uuid.New().String(),
		User: &pb.User{Name: "Test User", Email: "test@example.com"},
	}
	validUUID := uuid.MustParse(req.Uid)
	protoUser := gutils.ProtoToModel(req.User)
	expectedUser := &md.User{}

	// Case 1: Unauthorized
	t.Run("Unauthorized", func(t *testing.T) {
		invalidCtx := context.Background() // No "uid" in the context
		res, err := h.UpdateUser(invalidCtx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	// Case 2: Invalid UUID
	t.Run("Invalid UUID", func(t *testing.T) {
		invalidReq := &pb.UserWithUid{Uid: "invalid-uuid"}
		res, err := h.UpdateUser(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Validation error
	t.Run("Validation error", func(t *testing.T) {
		invalidUserReq := &pb.UserWithUid{
			Uid:  req.Uid,
			User: &pb.User{Name: ""}, // Invalid user data
		}
		res, err := h.UpdateUser(ctx, invalidUserReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 4: Success case - user updated
	t.Run("User updated", func(t *testing.T) {
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), validUUID, protoUser).Return(expectedUser, nil).Times(1)

		expectedProtoUser := gutils.ModelToProto(expectedUser)

		res, err := h.UpdateUser(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedProtoUser, res)
	})

	// Case 5: User not found
	t.Run("User not found", func(t *testing.T) {
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), validUUID, protoUser).Return(nil, ctrl.ErrNotFound).Times(1)

		res, err := h.UpdateUser(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// Case 6: Internal error
	t.Run("Internal error", func(t *testing.T) {
		mockCtrl.EXPECT().UpdateUser(gomock.Any(), validUUID, protoUser).Return(nil, errors.New("internal error")).Times(1)

		res, err := h.UpdateUser(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestDeleteUser(t *testing.T) {
	ctrlMock := gomock.NewController(t)
	defer ctrlMock.Finish()

	mockCtrl := mocks.NewMockCtrl(ctrlMock)
	auth := m2.NewMockAuth(ctrlMock)
	h := New(auth, mockCtrl)

	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.UuidMsg{Uuid: "123e4567-e89b-12d3-a456-426614174000"}
	validUUID := uuid.MustParse(req.Uuid)

	// Case 1: Unauthorized
	t.Run("Unauthorized", func(t *testing.T) {
		invalidCtx := context.Background()
		res, err := h.DeleteUser(invalidCtx, req)

		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	// Case 2: Invalid UUID
	t.Run("Invalid UUID", func(t *testing.T) {
		invalidReq := &pb.UuidMsg{Uuid: "invalid-uuid"}
		res, err := h.DeleteUser(ctx, invalidReq)

		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	// Case 3: Success case - user deleted
	t.Run("User deleted", func(t *testing.T) {
		mockCtrl.EXPECT().DeleteUser(gomock.Any(), validUUID).Return(nil).Times(1)

		res, err := h.DeleteUser(ctx, req)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	})

	// Case 4: Internal error
	t.Run("Internal error", func(t *testing.T) {
		mockCtrl.EXPECT().DeleteUser(gomock.Any(), validUUID).Return(errors.New("internal error")).Times(1)

		res, err := h.DeleteUser(ctx, req)
		assert.Nil(t, res)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
