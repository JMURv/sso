package grpc

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	md "github.com/JMURv/sso/pkg/model"
	"github.com/google/uuid"
	pm "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type Ctrl interface {
	ParseClaims(ctx context.Context, token string) (map[string]any, error)
	GetUserByToken(ctx context.Context, token string) (*md.User, error)
	SendSupportEmail(ctx context.Context, uid uuid.UUID, theme, text string) error
	CheckForgotPasswordEmail(ctx context.Context, password string, uid uuid.UUID, code int) error
	SendForgotPasswordEmail(ctx context.Context, email string) error
	SendLoginCode(ctx context.Context, email, password string) error
	CheckLoginCode(ctx context.Context, email string, code int) (string, string, error)

	IsUserExist(ctx context.Context, email string) (isExist bool, err error)
	SearchUser(ctx context.Context, query string, page, size int) (*md.PaginatedUser, error)
	ListUsers(ctx context.Context, page, size int) (*md.PaginatedUser, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error)
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, u *md.User, fileName string, bytes []byte) (uuid.UUID, string, string, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *md.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	ListPermissions(ctx context.Context, page, size int) (*md.PaginatedPermission, error)
	GetPermission(ctx context.Context, id uint64) (*md.Permission, error)
	CreatePerm(ctx context.Context, req *md.Permission) (uint64, error)
	UpdatePerm(ctx context.Context, id uint64, req *md.Permission) error
	DeletePerm(ctx context.Context, id uint64) error
}

type Handler struct {
	pb.SSOServer
	pb.UsersServer
	pb.PermissionSvcServer
	srv  *grpc.Server
	hsrv *health.Server
	ctrl Ctrl
}

func New(auth ctrl.AuthService, ctrl Ctrl) *Handler {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			AuthUnaryInterceptor(auth),
			metrics.SrvMetrics.UnaryServerInterceptor(pm.WithExemplarFromContext(metrics.Exemplar)),
		),
		grpc.ChainStreamInterceptor(
			metrics.SrvMetrics.StreamServerInterceptor(pm.WithExemplarFromContext(metrics.Exemplar)),
		),
	)

	hsrv := health.NewServer()
	hsrv.SetServingStatus("sso", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(srv)
	return &Handler{
		ctrl: ctrl,
		srv:  srv,
		hsrv: hsrv,
	}
}

func (h *Handler) Start(port int) {
	pb.RegisterSSOServer(h.srv, h)
	pb.RegisterUsersServer(h.srv, h)
	pb.RegisterPermissionSvcServer(h.srv, h)
	grpc_health_v1.RegisterHealthServer(h.srv, h.hsrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := h.srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatal(err)
	}
}

func (h *Handler) Close() error {
	h.srv.GracefulStop()
	return nil
}

func AuthUnaryInterceptor(auth ctrl.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Debug("missing metadata")
			return handler(ctx, req)
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			zap.L().Debug("missing authorization token")
			return handler(ctx, req)
		}

		tokenStr := authHeaders[0]
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		claims, err := auth.VerifyToken(tokenStr)
		if err != nil {
			zap.L().Debug("invalid token", zap.Error(err))
			return handler(ctx, req)
		}

		ctx = context.WithValue(ctx, "uid", claims["uid"])
		return handler(ctx, req)
	}
}
