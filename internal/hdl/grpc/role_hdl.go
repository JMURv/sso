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

func (h *Handler) ListRoles(ctx context.Context, req *gen.SSO_ListReq) (*gen.SSO_RoleList, error) {
	const op = "sso.ListRoles.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.ListRoles(ctx, int(page), int(size), map[string]any{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &gen.SSO_RoleList{
		Data:        utils.ListRolesToProtoFromPointer(res.Data),
		Count:       res.Count,
		TotalPages:  int64(res.TotalPages),
		CurrentPage: int64(res.CurrentPage),
		HasNextPage: res.HasNextPage,
	}, nil
}

func (h *Handler) GetRole(ctx context.Context, req *gen.SSO_Uint64Msg) (*gen.SSO_Role, error) {
	const op = "sso.GetRole.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
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
	const op = "sso.CreateRole.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to parse uid", zap.String("op", op))
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
	const op = "sso.UpdateRole.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.UpdateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := validation.V.Struct(r); err != nil {
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
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
	const op = "sso.DeleteRole.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to parse uid", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	if req == nil || req.Uint64 == 0 {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.DeleteRole(ctx, req.Uint64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &gen.SSO_Empty{}, nil
}
