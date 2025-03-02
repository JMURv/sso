package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	ctrl "github.com/JMURv/sso/internal/ctrl"
	md "github.com/JMURv/sso/internal/models"
	gutils "github.com/JMURv/sso/internal/models/mapper"
	"github.com/JMURv/sso/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestHandler_ListPermissions(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()
	page := uint64(1)
	size := uint64(10)
	expectedData := &md.PaginatedPermission{}

	t.Run(
		"Invalid request", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: 0, Size: 0}
			res, err := h.ListPermissions(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: page, Size: size}

			mctrl.EXPECT().ListPermissions(gomock.Any(), int(page), int(size)).Return(expectedData, nil).Times(1)

			res, err := h.ListPermissions(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, expectedData.Count, res.Count)
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			req := &pb.SSO_ListReq{Page: page, Size: size}

			mctrl.EXPECT().ListPermissions(gomock.Any(), int(page), int(size)).Return(
				nil,
				errors.New("internal error"),
			).Times(1)

			res, err := h.ListPermissions(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestHandler_GetPermission(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()

	validID := uint64(1)
	invalidID := uint64(0)
	expectedData := &md.Permission{}

	t.Run(
		"Invalid ID", func(t *testing.T) {
			req := &pb.SSO_Uint64Msg{Uint64: invalidID}
			res, err := h.GetPermission(ctx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			req := &pb.SSO_Uint64Msg{Uint64: validID}
			mctrl.EXPECT().GetPermission(gomock.Any(), validID).Return(expectedData, nil).Times(1)

			res, err := h.GetPermission(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
		},
	)

	t.Run(
		"Not Found", func(t *testing.T) {
			req := &pb.SSO_Uint64Msg{Uint64: validID}
			mctrl.EXPECT().GetPermission(gomock.Any(), validID).Return(nil, ctrl.ErrNotFound).Times(1)

			res, err := h.GetPermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal Error", func(t *testing.T) {
			req := &pb.SSO_Uint64Msg{Uint64: validID}
			mctrl.EXPECT().GetPermission(gomock.Any(), validID).Return(nil, errors.New("internal error")).Times(1)

			res, err := h.GetPermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestHandler_CreatePermission(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	ctx := context.Background()
	req := &pb.SSO_Permission{
		Name: "Test Perm",
	}

	proto := &md.Permission{
		Name: req.Name,
	}

	idx := uint64(1)
	invalidIDX := uint64(0)
	t.Run(
		"Validation error", func(t *testing.T) {
			invalidReq := &pb.SSO_Permission{Name: ""}
			res, err := h.CreatePermission(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().CreatePerm(gomock.Any(), proto).
				Return(idx, nil).Times(1)

			res, err := h.CreatePermission(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, idx, res.Uint64)
		},
	)

	t.Run(
		"Already exists", func(t *testing.T) {
			mctrl.EXPECT().CreatePerm(gomock.Any(), proto).
				Return(invalidIDX, ctrl.ErrAlreadyExists).Times(1)

			res, err := h.CreatePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			mctrl.EXPECT().CreatePerm(gomock.Any(), proto).
				Return(invalidIDX, errors.New("internal error")).Times(1)

			res, err := h.CreatePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestHandler_UpdatePermission(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	idx := uint64(1)
	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.SSO_Permission{
		Id:   idx,
		Name: "new-name",
	}

	proto := gutils.PermissionFromProto(req)
	t.Run(
		"Unauthorized", func(t *testing.T) {
			invalidCtx := context.Background()
			res, err := h.UpdatePermission(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid ID", func(t *testing.T) {
			invalidReq := &pb.SSO_Permission{Id: uint64(0)}
			res, err := h.UpdatePermission(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Validation error", func(t *testing.T) {
			invalidReq := &pb.SSO_Permission{Id: idx, Name: ""}
			res, err := h.UpdatePermission(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Updated", func(t *testing.T) {
			mctrl.EXPECT().UpdatePerm(gomock.Any(), idx, proto).Return(nil).Times(1)

			res, err := h.UpdatePermission(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, &pb.SSO_Empty{}, res)
		},
	)

	t.Run(
		"Not found", func(t *testing.T) {
			mctrl.EXPECT().UpdatePerm(gomock.Any(), idx, proto).Return(ctrl.ErrNotFound).Times(1)

			res, err := h.UpdatePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			mctrl.EXPECT().UpdatePerm(
				gomock.Any(),
				idx,
				proto,
			).Return(errors.New("internal error")).Times(1)

			res, err := h.UpdatePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}

func TestHandler_DeletePermission(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockCtrl(mock)
	auth := mocks.NewMockAuthService(mock)
	h := New(auth, mctrl)

	idx := uint64(1)
	ctx := context.WithValue(context.Background(), "uid", "user-uid")
	req := &pb.SSO_Uint64Msg{Uint64: idx}

	t.Run(
		"Unauthorized", func(t *testing.T) {
			invalidCtx := context.Background()
			res, err := h.DeletePermission(invalidCtx, req)

			assert.Nil(t, res)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		},
	)

	t.Run(
		"Invalid ID", func(t *testing.T) {
			invalidReq := &pb.SSO_Uint64Msg{Uint64: uint64(0)}
			res, err := h.DeletePermission(ctx, invalidReq)

			assert.Nil(t, res)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		},
	)

	t.Run(
		"Success", func(t *testing.T) {
			mctrl.EXPECT().DeletePerm(gomock.Any(), idx).Return(nil).Times(1)

			res, err := h.DeletePermission(ctx, req)
			assert.Nil(t, err)
			assert.NotNil(t, res)
		},
	)

	t.Run(
		"Not found", func(t *testing.T) {
			mctrl.EXPECT().DeletePerm(gomock.Any(), idx).Return(ctrl.ErrNotFound).Times(1)

			res, err := h.DeletePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.NotFound, status.Code(err))
		},
	)

	t.Run(
		"Internal error", func(t *testing.T) {
			mctrl.EXPECT().DeletePerm(gomock.Any(), idx).Return(errors.New("internal error")).Times(1)

			res, err := h.DeletePermission(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, codes.Internal, status.Code(err))
		},
	)
}
