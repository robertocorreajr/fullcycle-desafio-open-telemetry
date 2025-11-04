package otel

import (
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/zipkin"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Config struct {
    ServiceName string
    ZipkinURL   string
}

func InitTracer(cfg Config) (*sdktrace.TracerProvider, error) {
    if cfg.ZipkinURL == "" {
        return nil, fmt.Errorf("missing zipkin url")
    }
    exporter, err := zipkin.New(cfg.ZipkinURL)
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(cfg.ServiceName),
        )),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
