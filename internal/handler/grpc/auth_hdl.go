package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	"github.com/JMURv/sso/internal/validation"
	utils "github.com/JMURv/sso/pkg/utils/grpc"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

func (h *Handler) GetUserByToken(ctx context.Context, req *pb.StringSSOMsg) (*pb.User, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.GetUserByToken.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	token := req.String_
	if req == nil || token == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	u, err := h.ctrl.GetUserByToken(ctx, token)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return utils.ModelToProto(u), nil
}

func (h *Handler) ValidateToken(ctx context.Context, req *pb.StringSSOMsg) (*pb.BoolSSOMsg, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.ValidateToken.hdl"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	token := req.GetString_()
	if req == nil || token == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	return &pb.BoolSSOMsg{Bool: h.ctrl.ValidateToken(ctx, token)}, nil
}

func (h *Handler) SendLoginCode(ctx context.Context, req *pb.SendLoginCodeReq) (*pb.EmptySSO, error) {
	const op = "sso.SendLoginCode.hdl"
	s, c := time.Now(), codes.OK

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if req == nil || req.Email == "" || req.Password == "" {
		c = codes.InvalidArgument
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	if err := validation.ValidateEmail(req.Email); err != nil {
		c = codes.InvalidArgument
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.SendLoginCode(ctx, req.Email, req.Password)
	if err != nil && errors.Is(err, ctrl.ErrInvalidCredentials) {
		c = codes.InvalidArgument
		zap.L().Debug("failed to send login code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		span.SetTag("error", true)
		c = codes.Internal
		zap.L().Error("failed to send login code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}
	return &pb.EmptySSO{}, nil
}

func (h *Handler) CheckLoginCode(ctx context.Context, req *pb.CheckLoginCodeReq) (*pb.CheckLoginCodeRes, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.checkLoginCode.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	email, code := req.Email, req.Code
	if email == "" || code == 0 {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	if err := validation.ValidateEmail(email); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	access, refresh, err := h.ctrl.CheckLoginCode(ctx, email, int(code))
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to check login code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.CheckLoginCodeRes{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (h *Handler) CheckEmail(ctx context.Context, req *pb.EmailMsg) (*pb.CheckEmailRes, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.CheckEmail.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if req == nil || req.Email == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	isExist, err := h.ctrl.IsUserExist(ctx, req.Email)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to check email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.CheckEmailRes{
		IsExist: isExist,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, _ *pb.EmptySSO) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.Logout.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(uidStr)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	return &pb.EmptySSO{}, nil
}

func (h *Handler) SendForgotPasswordEmail(ctx context.Context, req *pb.EmailMsg) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.SendForgotPasswordEmail.handler"
	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if req == nil || req.Email == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	err := h.ctrl.SendForgotPasswordEmail(ctx, req.Email)
	if err != nil && errors.Is(err, ctrl.ErrInvalidCredentials) {
		c = codes.InvalidArgument
		zap.L().Debug("invalid credentials", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to send forgot password email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.EmptySSO{}, nil
}

func (h *Handler) CheckForgotPasswordEmail(ctx context.Context, req *pb.CheckForgotPasswordEmailReq) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.CheckForgotPasswordEmail.handler"
	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	pass, uid, token := req.Password, req.Uidb64, req.Token
	if pass == "" || uid == "" || token == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	uidb64, err := uuid.Parse(req.Uidb64)
	if err != nil {
		c = codes.InvalidArgument
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	intToken, err := strconv.Atoi(req.Token)
	if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	err = h.ctrl.CheckForgotPasswordEmail(ctx, pass, uidb64, intToken)
	if err != nil && errors.Is(err, ctrl.ErrCodeIsNotValid) {
		c = codes.InvalidArgument
		zap.L().Debug("invalid code", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("failed to find user", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrNotFound.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to check forgot password email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.EmptySSO{}, nil
}

func (h *Handler) SendSupportEmail(ctx context.Context, req *pb.SendSupportEmailReq) (*pb.EmptySSO, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.SendSupportEmail.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(uidStr)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	if req == nil || req.Theme == "" || req.Text == "" {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	err = h.ctrl.SendSupportEmail(ctx, uid, req.Theme, req.Text)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("failed to find user", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to send email", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, "failed to send email")
	}

	return &pb.EmptySSO{}, nil
}

func (h *Handler) Me(ctx context.Context, _ *pb.EmptySSO) (*pb.User, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.Me.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(uidStr)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	u, err := h.ctrl.GetUserByID(ctx, uid)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to get user", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return utils.ModelToProto(u), nil
}

func (h *Handler) UpdateMe(ctx context.Context, req *pb.User) (*pb.User, error) {
	s, c := time.Now(), codes.OK
	const op = "sso.UpdateMe.handler"

	span := opentracing.GlobalTracer().StartSpan(op)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	uidStr, ok := ctx.Value("uid").(string)
	if !ok {
		c = codes.Unauthenticated
		zap.L().Debug("failed to get uid from context", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrUnauthorized.Error())
	}

	uid, err := uuid.Parse(uidStr)
	if uid == uuid.Nil || err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to parse uid", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrParseUUID.Error())
	}

	if req == nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	user := utils.ProtoToModel(req)
	if err = validation.UserValidation(user); err != nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to validate obj", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	}

	err = h.ctrl.UpdateUser(ctx, uid, user)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		zap.L().Debug("user not found", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to update user", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.User{}, nil
}
