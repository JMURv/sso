package main

import (
	"context"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache/redis"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/hdl/grpc"
	"github.com/JMURv/sso/internal/hdl/http"
	"github.com/JMURv/sso/internal/observability/metrics/prometheus"
	"github.com/JMURv/sso/internal/observability/tracing/jaeger"
	"github.com/JMURv/sso/internal/repo/db"
	"github.com/JMURv/sso/internal/repo/s3"
	"github.com/JMURv/sso/internal/smtp"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "configs/config.yaml"

func mustRegisterLogger(mode string) {
	switch mode {
	case "prod":
		zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
	case "dev":
		zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	}
}

// TODO: fix caching
func main() {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Panic("unexpected panic occurred", zap.Any("error", err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.MustLoad(configPath)
	mustRegisterLogger(conf.Mode)

	go prometheus.New(conf.Prometheus.Port).Start(ctx)
	go jaeger.Start(ctx, conf.ServiceName, conf)

	au := auth.New(conf)
	cache := redis.New(conf)
	repo := db.New(conf)
	svc := ctrl.New(repo, au, cache, s3.New(conf), smtp.New(conf))
	h := http.New(svc, au)
	hg := grpc.New(conf.ServiceName, svc, au)

	go h.Start(conf.Server.Port)
	go hg.Start(conf.Server.GRPCPort)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	zap.L().Info("Shutting down gracefully...")
	if err := h.Close(ctx); err != nil {
		zap.L().Warn("Error closing handler", zap.Error(err))
	}

	if err := hg.Close(); err != nil {
		zap.L().Warn("Error closing grpc handler", zap.Error(err))
	}

	if err := cache.Close(); err != nil {
		zap.L().Warn("Failed to close connection to cache: ", zap.Error(err))
	}

	if err := repo.Close(); err != nil {
		zap.L().Warn("Error closing repository", zap.Error(err))
	}
}
