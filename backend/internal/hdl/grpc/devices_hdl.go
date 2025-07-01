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
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ListDevices(ctx context.Context, req *pb.SSO_ListDevicesRequest) (*pb.SSO_ListDevicesResponse, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.ListDevices(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &pb.SSO_ListDevicesResponse{
		Data: utils.ListDevicesToProto(res),
	}, nil
}

func (h *Handler) GetDevice(ctx context.Context, req *pb.SSO_StringMsg) (*pb.SSO_Device, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	res, err := h.ctrl.GetDevice(ctx, uid, req.String_)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return utils.DeviceToProto(res), nil
}

func (h *Handler) UpdateDevice(ctx context.Context, req *pb.SSO_UpdateDeviceRequest) (*pb.SSO_Empty, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to get uid from context")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	if req == nil || req.Id == "" {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	r := &dto.UpdateDeviceRequest{Name: req.Name}
	if err := validation.V.Struct(r); err != nil {
		zap.L().Error("failed to validate obj", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := h.ctrl.UpdateDevice(ctx, uid, req.Id, r)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}

func (h *Handler) DeleteDevice(ctx context.Context, req *pb.SSO_StringMsg) (*pb.SSO_Empty, error) {
	uid, ok := ctx.Value("uid").(uuid.UUID)
	if !ok {
		zap.L().Error("failed to parse uid")
		return nil, status.Errorf(codes.InvalidArgument, ctrl.ErrParseUUID.Error())
	}

	if req == nil || req.String_ == "" {
		zap.L().Error("failed to decode request")
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.DeleteDevice(ctx, uid, req.String_)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}
	return &pb.SSO_Empty{}, nil
}
