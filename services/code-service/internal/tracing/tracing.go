// Package tracing initialise OpenTelemetry pour exporter les traces vers Tempo (OTLP/gRPC).
package tracing

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init configure le TracerProvider global avec un exporteur OTLP/gRPC vers endpoint (ex: "tempo:4317").
// Retourne une fonction shutdown à appeler à l'arrêt du service.
func Init(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// SkipHealthAndMetrics indique à otelgin de ne pas tracer les endpoints
// de healthcheck et de scraping Prometheus, trop fréquents pour être utiles dans Tempo.
func SkipHealthAndMetrics(r *http.Request) bool {
	switch r.URL.Path {
	case "/health", "/healthz", "/metrics":
		return false
	}
	return true
}
