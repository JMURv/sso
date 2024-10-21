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
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (h *Handler) SearchUser(ctx context.Context, req *pb.SearchReq) (*pb.PaginatedUsersRes, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.SearchUser.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	q, page, size := req.Query, req.Page, req.Size
	if q == "" || page == 0 || size == 0 {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	u, err := h.ctrl.SearchUser(ctx, q, int(page), int(size))
	if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.PaginatedUsersRes{
		Data:        utils.ListModelToProto(u.Data),
		Count:       u.Count,
		TotalPages:  int64(u.TotalPages),
		CurrentPage: int64(u.CurrentPage),
		HasNextPage: u.HasNextPage,
	}, nil
}

func (h *Handler) ListUsers(ctx context.Context, req *pb.ListReq) (*pb.PaginatedUsersRes, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.ListUsers.hdl"

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

	u, err := h.ctrl.ListUsers(ctx, int(page), int(size))
	if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.PaginatedUsersRes{
		Data:        utils.ListModelToProto(u.Data),
		Count:       u.Count,
		TotalPages:  int64(u.TotalPages),
		CurrentPage: int64(u.CurrentPage),
		HasNextPage: u.HasNextPage,
	}, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.CreateUser.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	protoUser := &md.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := validation.NewUserValidation(protoUser); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate user", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	uid, access, refresh, err := h.ctrl.CreateUser(ctx, protoUser, req.File.Filename, req.File.File)
	if err != nil && errors.Is(err, ctrl.ErrAlreadyExists) {
		c = codes.AlreadyExists
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.CreateUserRes{
		Uid:     uid.String(),
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (h *Handler) GetUser(ctx context.Context, req *pb.UuidMsg) (*pb.User, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.GetUser.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	uid, err := uuid.Parse(req.Uuid)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	u, err := h.ctrl.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return utils.ModelToProto(u), nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *pb.UserWithUid) (*pb.UuidMsg, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.UpdateUser.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	_, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(req.Uid)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	protoUser := utils.ProtoToModel(req.User)
	if err = validation.UserValidation(protoUser); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	err = h.ctrl.UpdateUser(ctx, uid, protoUser)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.UuidMsg{Uuid: uid.String()}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, req *pb.UuidMsg) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.DeleteUser.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	_, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(req.Uuid)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	err = h.ctrl.DeleteUser(ctx, uid)
	if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.EmptySSO{}, nil
}
