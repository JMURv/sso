package main

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/cache/redis"
	ctrl "github.com/JMURv/sso/internal/controller"
	discovery "github.com/JMURv/sso/internal/discovery/JMURv/grpc"
	//handler "github.com/JMURv/sso/internal/handler/http"
	handler "github.com/JMURv/sso/internal/handler/grpc"
	tracing "github.com/JMURv/sso/internal/metrics/jaeger"
	metrics "github.com/JMURv/sso/internal/metrics/prometheus"
	db "github.com/JMURv/sso/internal/repository/db"
	"github.com/JMURv/sso/internal/smtp"
	cfg "github.com/JMURv/sso/pkg/config"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "local.config.yaml"

func mustRegisterLogger(mode string) {
	switch mode {
	case "prod":
		zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
	case "dev":
		zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Panic("panic occurred", zap.Any("error", err))
			os.Exit(1)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	conf := cfg.MustLoad(configPath)
	mustRegisterLogger(conf.Server.Mode)

	// Start metrics and tracing
	go metrics.New(conf.Server.Port + 5).Start(ctx)
	go tracing.Start(ctx, conf.ServiceName, conf.Jaeger)

	dscvry := discovery.New(
		conf.SrvDiscovery.URL,
		conf.ServiceName,
		fmt.Sprintf("%v://%v:%v", conf.Server.Scheme, conf.Server.Domain, conf.Server.Port),
	)

	if err := dscvry.Register(ctx); err != nil {
		zap.L().Warn("Error registering service", zap.Error(err))
	}

	// Setting up main app
	cache := redis.New(conf.Redis)
	repo := db.New(conf.DB)
	email := smtp.New(conf.Email, conf.Server)

	au := auth.New(conf.Auth.Secret)
	svc := ctrl.New(au, repo, cache, email)
	h := handler.New(au, svc)

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-c

		zap.L().Info("Shutting down gracefully...")

		cancel()
		if err := cache.Close(); err != nil {
			zap.L().Debug("Failed to close connection to Redis: ", zap.Error(err))
		}

		if err := dscvry.Deregister(ctx); err != nil {
			zap.L().Warn("Error deregistering from discovery", zap.Error(err))
		}

		if err := dscvry.Close(); err != nil {
			zap.L().Warn("Error closing discovery", zap.Error(err))
		}

		if err := h.Close(); err != nil {
			zap.L().Warn("Error closing handler", zap.Error(err))
		}

		os.Exit(0)
	}()

	// Start service
	zap.L().Info(
		fmt.Sprintf("Starting server on %v://%v:%v", conf.Server.Scheme, conf.Server.Domain, conf.Server.Port),
	)
	h.Start(conf.Server.Port)
}
