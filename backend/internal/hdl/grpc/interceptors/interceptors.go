package interceptors

import (
	"context"
	"github.com/JMURv/sso/internal/auth"
	metrics "github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

func Auth(au auth.Core) grpc.UnaryServerInterceptor {
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

		claims, err := au.ParseClaims(ctx, tokenStr)
		if err != nil {
			return handler(ctx, req)
		}

		ctx = context.WithValue(ctx, "uid", claims.UID)
		return handler(ctx, req)
	}
}

func LogTraceMetrics() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		s := time.Now()
		span, ctx := opentracing.StartSpanFromContext(ctx, info.FullMethod)
		defer span.Finish()

		res, err := handler(ctx, req)
		statusCode := status.Code(err)
		metrics.ObserveRequest(time.Since(s), int(statusCode), info.FullMethod)

		zap.L().Info(
			"<--",
			zap.String("method", info.FullMethod),
			zap.Int("status", int(statusCode)),
			zap.Any("duration", time.Since(s)),
			zap.Error(err),
		)
		return res, err
	}
}

func Device() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			zap.L().Debug("missing metadata")
			return handler(ctx, req)
		}

		ua := md["User-Agent"]
		if len(ua) == 0 {
			zap.L().Error("missing user agent header")
			return handler(ctx, req)
		}

		ip := md["IP"]
		if len(ua) == 0 {
			zap.L().Error("missing ip header")
			return handler(ctx, req)
		}

		ctx = context.WithValue(ctx, "ua", ua)
		ctx = context.WithValue(ctx, "ip", ip)
		return handler(ctx, req)
	}
}
