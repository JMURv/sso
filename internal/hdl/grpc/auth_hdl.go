package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/grpc/utils"
	"github.com/JMURv/sso/internal/hdl/validation"
	"github.com/JMURv/sso/internal/models/mapper"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

func (h *Handler) Authenticate(ctx context.Context, req *pb.SSO_EmailAndPasswordRequest) (*pb.SSO_TokenPair, error) {
	const op = "sso.Authenticate.hdl"
	if req == nil {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	d := utils.ParseDeviceFromContext(ctx)
	r := &dto.EmailAndPasswordRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.Authenticate(ctx, &d, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_TokenPair{
		Access:  res.Access,
		Refresh: res.Refresh,
	}, nil
}

func (h *Handler) ParseClaims(ctx context.Context, req *pb.SSO_StringMsg) (*pb.SSO_ParseClaimsRes, error) {
	const op = "sso.ParseClaims.hdl"
	token := req.GetString_()
	if req == nil || token == "" {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.ParseClaims(ctx, token)
	if err != nil {
		zap.L().Debug("failed to parse claims", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_ParseClaimsRes{
		Uid:   res.UID.String(),
		Roles: mapper.ListRolesToProto(res.Roles),
		Exp:   res.ExpiresAt.UnixMilli(),
		Iat:   res.IssuedAt.UnixMilli(),
		Sub:   res.Subject,
	}, nil
}

func (h *Handler) Refresh(ctx context.Context, req *pb.SSO_RefreshRequest) (*pb.SSO_TokenPair, error) {
	const op = "sso.Authenticate.hdl"
	if req == nil {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	d := utils.ParseDeviceFromContext(ctx)
	r := &dto.RefreshRequest{Refresh: req.Refresh}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.Refresh(ctx, &d, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_TokenPair{
		Access:  res.Access,
		Refresh: res.Refresh,
	}, nil
}

func (h *Handler) SendLoginCode(ctx context.Context, req *pb.SSO_SendLoginCodeReq) (*pb.SSO_TokenPair, error) {
	const op = "sso.SendLoginCode.hdl"
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	d := utils.ParseDeviceFromContext(ctx)
	r := &dto.LoginCodeRequest{Email: req.Email, Password: req.Password}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.SendLoginCode(ctx, &d, r.Email, r.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		zap.L().Error("failed to send login code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_TokenPair{
		Access:  res.Access,
		Refresh: res.Refresh,
	}, nil
}

func (h *Handler) CheckLoginCode(ctx context.Context, req *pb.SSO_CheckLoginCodeReq) (*pb.SSO_TokenPair, error) {
	const op = "sso.checkLoginCode.handler"

	email, code := req.Email, req.Code
	if email == "" || code == 0 {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	d := utils.ParseDeviceFromContext(ctx)
	r := &dto.CheckLoginCodeRequest{Email: req.Email, Code: int(req.Code)}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.ctrl.CheckLoginCode(ctx, &d, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		zap.L().Error("failed to check login code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_TokenPair{
		Access:  res.Access,
		Refresh: res.Refresh,
	}, nil
}

func (h *Handler) SendForgotPasswordEmail(ctx context.Context, req *pb.SSO_EmailMsg) (*pb.SSO_Empty, error) {
	const op = "sso.SendForgotPasswordEmail.handler"
	if req == nil || req.Email == "" {
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	r := dto.SendForgotPasswordEmail{}
	if err := validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.SendForgotPasswordEmail(ctx, r.Email)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		zap.L().Error("failed to send forgot password email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}

func (h *Handler) CheckForgotPasswordEmail(ctx context.Context, req *pb.SSO_CheckForgotPasswordEmailReq) (*pb.SSO_Empty, error) {
	const op = "sso.CheckForgotPasswordEmail.handler"

	uid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	code, err := strconv.Atoi(req.Code)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	r := &dto.CheckForgotPasswordEmailRequest{Password: req.Password, ID: uid, Code: code}
	if err = validation.V.Struct(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = h.ctrl.CheckForgotPasswordEmail(ctx, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrCodeIsNotValid) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, ctrl.ErrNotFound.Error())
		}
		zap.L().Debug("failed to check forgot password email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_Empty{}, nil
}

func (h *Handler) Logout(ctx context.Context, _ *pb.SSO_Empty) (*pb.SSO_Empty, error) {
	const op = "sso.Logout.handler"
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(codes.Unauthenticated, hdl.ErrFailedToParseUUID.Error())
	}

	err := h.ctrl.Logout(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}
