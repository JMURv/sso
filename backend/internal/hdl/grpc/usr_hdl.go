package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/validation"
	utils "github.com/JMURv/sso/internal/models/mapper"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ExistUser(ctx context.Context, req *pb.SSO_ExistUserRequest) (*pb.SSO_ExistUserResponse, error) {
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

func (h *Handler) GetMe(ctx context.Context, _ *pb.SSO_Empty) (*pb.SSO_User, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
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

func (h *Handler) ListUsers(ctx context.Context, req *pb.SSO_UserListRequest) (*pb.SSO_UserListResponse, error) {
	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	u, err := h.ctrl.ListUsers(ctx, int(page), int(size), map[string]any{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_UserListResponse{
		Data:        utils.ListModelToProto(u.Data),
		Count:       u.Count,
		TotalPages:  int64(u.TotalPages),
		CurrentPage: int64(u.CurrentPage),
		HasNextPage: u.HasNextPage,
	}, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *pb.SSO_CreateUserReq) (*pb.SSO_CreateUserRes, error) {
	if _, ok := ctx.Value("uid").(uuid.UUID); !ok {
		zap.L().Error("failed to parse uid")
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

	res, err := h.ctrl.CreateUser(ctx, r, &s3.UploadFileRequest{})
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
	uid, err := uuid.Parse(req.Uuid)
	if uid == uuid.Nil || err != nil {
		zap.L().Error("failed to parse uid", zap.Error(err))
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
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
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
		zap.L().Error("failed to validate obj", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.UpdateUser(ctx, uid, r, &s3.UploadFileRequest{})
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_UuidMsg{Uuid: uid.String()}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, _ *pb.SSO_UuidMsg) (*pb.SSO_Empty, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to parse uid")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	err := h.ctrl.DeleteUser(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}
