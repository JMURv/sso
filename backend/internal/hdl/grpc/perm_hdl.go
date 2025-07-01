package grpc

import (
	"context"
	"errors"

	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/validation"
	utils "github.com/JMURv/sso/internal/models/mapper"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ListPermissions(ctx context.Context, req *pb.SSO_PermissionListRequest) (*pb.SSO_PermissionListResponse, error) {
	page := req.Page
	if page < 1 {
		page = config.DefaultPage
	}

	size := req.Size
	if size < 1 {
		size = config.DefaultSize
	}

	res, err := h.ctrl.ListPermissions(ctx, int(page), int(size), map[string]any{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_PermissionListResponse{
		Data:        utils.ListPermissionsToProto(res.Data),
		Count:       res.Count,
		TotalPages:  int64(res.TotalPages),
		CurrentPage: int64(res.CurrentPage),
		HasNextPage: res.HasNextPage,
	}, nil
}

func (h *Handler) GetPermission(ctx context.Context, req *pb.SSO_Uint64Msg) (*pb.SSO_Permission, error) {
	if req == nil || req.Uint64 == 0 {
		zap.L().Error("failed to parse uid", zap.Uint64("uid", req.Uint64))

		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.GetPermission(ctx, req.Uint64)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return utils.PermissionToProto(res), nil
}

func (h *Handler) CreatePermission(ctx context.Context, req *pb.SSO_Permission) (*pb.SSO_Uint64Msg, error) {
	mdPerm := &dto.CreatePermissionRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := validation.V.Struct(mdPerm); err != nil {
		zap.L().Error("failed to validate obj", zap.Error(err))

		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	uid, err := h.ctrl.CreatePerm(ctx, mdPerm)
	if err != nil {
		if errors.Is(err, ctrl.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}

		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_Uint64Msg{
		Uint64: uid,
	}, nil
}

func (h *Handler) UpdatePermission(ctx context.Context, req *pb.SSO_Permission) (*pb.SSO_Empty, error) {
	if _, ok := ctx.Value("uid").(uuid.UUID); !ok {
		zap.L().Error("failed to get uid from context")

		return nil, status.Errorf(codes.Unauthenticated, hdl.ErrFailedToParseUUID.Error())
	}

	if req == nil || req.Id == 0 {
		zap.L().Error("failed to decode request")

		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	mdPerm := &dto.UpdatePermissionRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := validation.V.Struct(mdPerm); err != nil {
		zap.L().Error("failed to validate obj", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.UpdatePerm(ctx, req.Id, mdPerm)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}

func (h *Handler) DeletePermission(ctx context.Context, req *pb.SSO_Uint64Msg) (*pb.SSO_Empty, error) {
	if _, ok := ctx.Value("uid").(uuid.UUID); !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.Unauthenticated, hdl.ErrFailedToParseUUID.Error())
	}

	if req == nil || req.Uint64 == 0 {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.DeletePerm(ctx, req.Uint64)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}
