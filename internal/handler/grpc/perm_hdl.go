package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/grpc"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (h *Handler) ListPermissions(ctx context.Context, req *pb.ListReq) (*pb.PermissionList, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.ListPermissions.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	page, size := req.Page, req.Size
	if page == 0 || size == 0 {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.ListPermissions(ctx, int(page), int(size))
	if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.PermissionList{
		Data:        utils.ListPermissionsToProto(res.Data),
		Count:       res.Count,
		TotalPages:  int64(res.TotalPages),
		CurrentPage: int64(res.CurrentPage),
		HasNextPage: res.HasNextPage,
	}, nil
}

func (h *Handler) GetPermission(ctx context.Context, req *pb.Uint64Msg) (*pb.Permission, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.GetPermission.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if req == nil || req.Uint64 == 0 {
		c = codes.InvalidArgument
		zap.L().Debug(
			"failed to parse uid",
			zap.String("op", op),
			zap.Uint64("uid", req.Uint64),
		)
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.GetPermission(ctx, req.Uint64)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return utils.PermissionToProto(res), nil
}

func (h *Handler) CreatePermission(ctx context.Context, req *pb.Permission) (*pb.Uint64Msg, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.CreatePermission.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	mdPerm := &md.Permission{
		ID:   req.Id,
		Name: req.Name,
	}

	if err := validation.PermValidation(mdPerm); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	uid, err := h.ctrl.CreatePerm(ctx, mdPerm)
	if err != nil && errors.Is(err, ctrl.ErrAlreadyExists) {
		c = codes.AlreadyExists
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.Uint64Msg{
		Uint64: uid,
	}, nil
}

func (h *Handler) UpdatePermission(ctx context.Context, req *pb.Permission) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.UpdatePermission.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if _, ok := ctx.Value("uid").(string); !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	if req == nil || req.Id == 0 || req.Name == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	mdPerm := utils.PermissionFromProto(req)
	if err := validation.PermValidation(mdPerm); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	if err := h.ctrl.UpdatePerm(ctx, req.Id, mdPerm); err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.EmptySSO{}, nil
}

func (h *Handler) DeletePermission(ctx context.Context, req *pb.Uint64Msg) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.DeletePermission.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if _, ok := ctx.Value("uid").(string); !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	if req == nil || req.Uint64 == 0 {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	if err := h.ctrl.DeletePerm(ctx, req.Uint64); err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.EmptySSO{}, nil
}
