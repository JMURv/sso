package grpc

import (
	"context"
	"errors"

	"github.com/JMURv/sso/api/grpc/v1/gen"
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

func (h *Handler) ListRoles(ctx context.Context, req *gen.SSO_RoleListRequest) (*gen.SSO_RoleListResponse, error) {
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.ListRoles(ctx, int(page), int(size), map[string]any{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &gen.SSO_RoleListResponse{
		Data:        utils.ListRolesToProtoFromPointer(res.Data),
		Count:       res.Count,
		TotalPages:  int64(res.TotalPages),
		CurrentPage: int64(res.CurrentPage),
		HasNextPage: res.HasNextPage,
	}, nil
}

func (h *Handler) GetRole(ctx context.Context, req *gen.SSO_Uint64Msg) (*gen.SSO_Role, error) {
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	res, err := h.ctrl.GetRole(ctx, req.Uint64)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return utils.RoleToProto(res), nil
}

func (h *Handler) CreateRole(ctx context.Context, req *gen.SSO_Role) (*gen.SSO_Uint64Msg, error) {
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to parse uid")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.CreateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.CreateRole(ctx, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &gen.SSO_Uint64Msg{
		Uint64: res,
	}, nil
}

func (h *Handler) UpdateRole(ctx context.Context, req *gen.SSO_Role) (*gen.SSO_Empty, error) {
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.UpdateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := validation.V.Struct(r); err != nil {
		zap.L().Error("failed to validate obj", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.UpdateRole(ctx, req.Id, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &gen.SSO_Empty{}, nil
}

func (h *Handler) DeleteRole(ctx context.Context, req *gen.SSO_Uint64Msg) (*gen.SSO_Empty, error) {
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to parse uid")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	if req == nil || req.Uint64 == 0 {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.DeleteRole(ctx, req.Uint64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &gen.SSO_Empty{}, nil
}
