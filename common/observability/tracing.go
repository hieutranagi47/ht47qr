package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func ConfigureTraceProvider() *tracesdk.TracerProvider {
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = fmt.Sprintf("%s/jaeger-api/v1/traces", os.Getenv("GATEWAY_ADDR"))
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "htqrcode"
	}

	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL(otlpEndpoint),
	)
	if err != nil {
		panic(err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	// Propagate trace context across service boundaries.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp
}
