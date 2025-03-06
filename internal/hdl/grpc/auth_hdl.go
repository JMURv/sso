package grpc

import (
	"context"
	"errors"
	pb "github.com/JMURv/sso/api/grpc/v1/gen"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl/validation"
	utils "github.com/JMURv/sso/internal/models/mapper"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/google/uuid"
	ot "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

func (h *Handler) Authenticate(ctx context.Context, req *pb.SSO_EmailAndPasswordRequest) (*pb.SSO_TokenPair, error) {
	const op = "sso.Authenticate.hdl"
	s, c := time.Now(), codes.OK
	span, ctx := ot.StartSpanFromContext(ctx, op)
	defer func() {
		span.Finish()
		metrics.ObserveRequest(time.Since(s), int(c), op)
	}()

	if req == nil {
		c = codes.InvalidArgument
		zap.L().Debug("failed to decode request", zap.String("op", op))
		return nil, status.Errorf(c, ctrl.ErrDecodeRequest.Error())
	}

	r := &dto.EmailAndPasswordRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	err := validation.LoginAndPasswordRequest(r)
	if err != nil {
		c = codes.InvalidArgument
		return nil, status.Errorf(c, err.Error())
	}

	res, err := h.ctrl.Authenticate(ctx, r)
	if err != nil && errors.Is(err, ctrl.ErrNotFound) {
		c = codes.NotFound
		return nil, status.Errorf(c, err.Error())
	} else if err != nil {
		c = codes.Internal
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.SSO_TokenPair{
		Token: res.Token,
	}, nil
}

func (h *Handler) GetUserByToken(ctx context.Context, req *pb.SSO_StringMsg) (*pb.SSO_User, error) {
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

func (h *Handler) ParseClaims(ctx context.Context, req *pb.SSO_StringMsg) (*pb.SSO_ParseClaimsRes, error) {
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

	res, err := h.ctrl.ParseClaims(ctx, token)
	if err != nil {
		c = codes.Internal
		zap.L().Debug("failed to parse claims", zap.String("op", op), zap.Error(err))
		return nil, status.Errorf(c, ctrl.ErrInternalError.Error())
	}

	return &pb.SSO_ParseClaimsRes{
		Token: res["uid"].(string),
		Email: res["email"].(string),
		Exp:   int64(res["exp"].(float64)),
	}, nil
}

func (h *Handler) SendLoginCode(ctx context.Context, req *pb.SSO_SendLoginCodeReq) (*pb.SSO_Empty, error) {
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
	return &pb.SSO_Empty{}, nil
}

func (h *Handler) CheckLoginCode(ctx context.Context, req *pb.SSO_CheckLoginCodeReq) (*pb.SSO_CheckLoginCodeRes, error) {
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

	return &pb.SSO_CheckLoginCodeRes{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (h *Handler) CheckEmail(ctx context.Context, req *pb.SSO_EmailMsg) (*pb.SSO_CheckEmailRes, error) {
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

	return &pb.SSO_CheckEmailRes{
		IsExist: isExist,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, _ *pb.SSO_Empty) (*pb.SSO_Empty, error) {
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

	return &pb.SSO_Empty{}, nil
}

func (h *Handler) SendForgotPasswordEmail(ctx context.Context, req *pb.SSO_EmailMsg) (*pb.SSO_Empty, error) {
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

	return &pb.SSO_Empty{}, nil
}

func (h *Handler) CheckForgotPasswordEmail(ctx context.Context, req *pb.SSO_CheckForgotPasswordEmailReq) (*pb.SSO_Empty, error) {
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

	return &pb.SSO_Empty{}, nil
}

func (h *Handler) SendSupportEmail(ctx context.Context, req *pb.SSO_SendSupportEmailReq) (*pb.SSO_Empty, error) {
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

	return &pb.SSO_Empty{}, nil
}

func (h *Handler) Me(ctx context.Context, _ *pb.SSO_Empty) (*pb.SSO_User, error) {
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

func (h *Handler) UpdateMe(ctx context.Context, req *pb.SSO_User) (*pb.SSO_User, error) {
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

	return &pb.SSO_User{}, nil
}
