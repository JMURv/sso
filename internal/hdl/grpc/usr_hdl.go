package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	ctrl "github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/validation"
	utils "github.com/JMURv/sso/internal/models/mapper"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ExistUser(ctx context.Context, req *pb.SSO_ExistUserRequest) (*pb.SSO_ExistUserResponse, error) {
	const op = "sso.ExistUser.hdl"
	res, err := h.ctrl.IsUserExist(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_ExistUserResponse{
		IsExist: res.Exists,
	}, nil
}

func (h *Handler) GetMe(ctx context.Context, req *pb.SSO_Empty) (*pb.SSO_User, error) {
	const op = "sso.GetMe.hdl"
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	res, err := h.ctrl.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return utils.ModelToProto(res), nil
}

func (h *Handler) ListUsers(ctx context.Context, req *pb.SSO_ListReq) (*pb.SSO_PaginatedUsersRes, error) {
	const op = "sso.ListUsers.hdl"

	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	u, err := h.ctrl.ListUsers(ctx, int(page), int(size), map[string]any{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_PaginatedUsersRes{
		Data:        utils.ListModelToProto(u.Data),
		Count:       u.Count,
		TotalPages:  int64(u.TotalPages),
		CurrentPage: int64(u.CurrentPage),
		HasNextPage: u.HasNextPage,
	}, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *pb.SSO_CreateUserReq) (*pb.SSO_CreateUserRes, error) {
	const op = "sso.CreateUser.hdl"
	_, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to parse uid", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		IsActive: req.IsActive,
		IsEmail:  req.IsEmailVerified,
		Roles:    req.Roles,
	}

	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.CreateUser(ctx, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_CreateUserRes{
		Uid: res.ID.String(),
	}, nil
}

func (h *Handler) GetUser(ctx context.Context, req *pb.SSO_UuidMsg) (*pb.SSO_User, error) {
	const op = "sso.GetUser.hdl"
	uid, err := uuid.Parse(req.Uuid)
	if uid == uuid.Nil || err != nil {
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	u, err := h.ctrl.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return utils.ModelToProto(u), nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *pb.SSO_UpdateUserReq) (*pb.SSO_UuidMsg, error) {
	const op = "sso.UpdateUser.hdl"
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.UpdateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		IsActive: req.IsActive,
		IsEmail:  req.IsEmailVerified,
		Roles:    req.Roles,
	}
	if err := validation.V.Struct(r); err != nil {
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.UpdateUser(ctx, uid, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_UuidMsg{Uuid: uid.String()}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *pb.SSO_UuidMsg) (*pb.SSO_Empty, error) {
	const op = "sso.DeleteUser.hdl"
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to parse uid", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	err := h.ctrl.DeleteUser(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}
