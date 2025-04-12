package grpc

import (
	"errors"
	"fmt"
	"github.com/JMURv/sso/api/grpc/v1/gen"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/hdl/grpc/interceptors"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	pm "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
)

type Handler struct {
	gen.AuthServer
	gen.UsersServer
	gen.DevicesServer
	gen.PermissionServer
	gen.RoleServer
	srv  *grpc.Server
	hsrv *health.Server
	ctrl ctrl.AppCtrl
	au   auth.Core
}

func New(name string, ctrl ctrl.AppCtrl, au auth.Core) *Handler {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.Auth(au),
			interceptors.Device(),
			interceptors.LogTraceMetrics(),
			metrics.SrvMetrics.UnaryServerInterceptor(
				pm.WithExemplarFromContext(metrics.Exemplar),
			),
		),
		grpc.ChainStreamInterceptor(
			metrics.SrvMetrics.StreamServerInterceptor(
				pm.WithExemplarFromContext(metrics.Exemplar),
			),
		),
	)

	reflection.Register(srv)

	hsrv := health.NewServer()
	hsrv.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
	return &Handler{
		ctrl: ctrl,
		srv:  srv,
		hsrv: hsrv,
		au:   au,
	}
}

func (h *Handler) Start(port int) {
	gen.RegisterAuthServer(h.srv, h)
	gen.RegisterUsersServer(h.srv, h)
	gen.RegisterDevicesServer(h.srv, h)
	gen.RegisterPermissionServer(h.srv, h)
	gen.RegisterRoleServer(h.srv, h)
	grpc_health_v1.RegisterHealthServer(h.srv, h.hsrv)

	portStr := fmt.Sprintf(":%v", port)
	lis, err := net.Listen("tcp", portStr)
	if err != nil {
		zap.L().Fatal("failed to listen", zap.Error(err))
	}

	zap.L().Info(
		"Starting GRPC server",
		zap.String("addr", portStr),
	)
	if err = h.srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		zap.L().Fatal("failed to serve", zap.Error(err))
	}
}

func (h *Handler) Close() error {
	h.srv.GracefulStop()
	return nil
}
