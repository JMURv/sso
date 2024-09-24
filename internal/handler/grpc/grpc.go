package grpc

import (
	"context"
	"fmt"
	pb "github.com/JMURv/sso/api/pb"
	ctrl "github.com/JMURv/sso/internal/controller"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	md "github.com/JMURv/sso/pkg/model"
	utils "github.com/JMURv/sso/pkg/utils/http"
	"github.com/google/uuid"
	pm "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type Ctrl interface {
	IsUserExist(ctx context.Context, email string) (isExist bool, err error)
	UserSearch(ctx context.Context, query string, page int, size int) (*utils.PaginatedData, error)
	ListUsers(ctx context.Context, page int, size int) (*utils.PaginatedData, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*md.User, error)
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, u *md.User, fileName string, bytes []byte) (*md.User, string, string, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, newData *md.User) (*md.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	SendSupportEmail(ctx context.Context, uid uuid.UUID, theme string, text string) error
	CheckForgotPasswordEmail(ctx context.Context, password string, uid uuid.UUID, code int) error
	SendForgotPasswordEmail(ctx context.Context, email string) error
	SendLoginCode(ctx context.Context, email string, password string) error
	CheckLoginCode(ctx context.Context, email string, code int) (string, string, error)
}

type Handler struct {
	pb.SSOServer
	pb.UsersServer
	srv  *grpc.Server
	ctrl Ctrl
}

func New(auth ctrl.Auth, ctrl Ctrl) *Handler {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			AuthUnaryInterceptor(auth),
			metrics.SrvMetrics.UnaryServerInterceptor(pm.WithExemplarFromContext(metrics.Exemplar)),
		),
		grpc.ChainStreamInterceptor(
			metrics.SrvMetrics.StreamServerInterceptor(pm.WithExemplarFromContext(metrics.Exemplar)),
		),
	)

	reflection.Register(srv)
	return &Handler{
		ctrl: ctrl,
		srv:  srv,
	}
}

func (h *Handler) Start(port int) {
	pb.RegisterSSOServer(h.srv, h)
	pb.RegisterUsersServer(h.srv, h)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Fatal(h.srv.Serve(lis))
}

func (h *Handler) Close() error {
	h.srv.GracefulStop()
	return nil
}

func AuthUnaryInterceptor(auth ctrl.Auth) grpc.UnaryServerInterceptor {
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
