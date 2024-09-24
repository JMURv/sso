package jaeger

import (
	"context"
	cfg "github.com/JMURv/sso/pkg/config"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"log"
)

func Start(ctx context.Context, serviceName string, conf *cfg.JaegerConfig) {
	tracerCfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  conf.Sampler.Type,
			Param: float64(conf.Sampler.Param),
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           conf.Reporter.LogSpans,
			LocalAgentHostPort: conf.Reporter.LocalAgentHostPort,
		},
	}

	tracer, closer, err := tracerCfg.NewTracer()
	if err != nil {
		log.Fatalf("Error initializing Jaeger tracer: %s", err.Error())
	}

	opentracing.SetGlobalTracer(tracer)

	zap.L().Debug("Jaeger has been started")
	<-ctx.Done()

	zap.L().Debug("Shutting down Jaeger")
	if err = closer.Close(); err != nil {
		zap.L().Debug("Error shutting down Jaeger", zap.Error(err))
	}
	zap.L().Debug("Jaeger has been stopped")
}
