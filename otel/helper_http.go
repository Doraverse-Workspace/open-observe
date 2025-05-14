package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func InitTracerHTTP(config Config) *sdktrace.TracerProvider {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	httpEndpoint := "127.0.0.1:5081"
	if config.Endpoint != "" {
		httpEndpoint = config.Endpoint
	}

	path := "/api/default/v1/traces"
	streamName := "default"
	if config.StreamName != "" {
		streamName = config.StreamName
	}

	fmt.Println("path", path)

	otlptracehttp.NewClient()

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(httpEndpoint),
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + config.BasicAuth,
			"stream-name":   streamName,
		}),
	}

	if !config.IsSecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	otlpHTTPExporter, err := otlptracehttp.New(context.TODO(), opts...)

	if err != nil {
		fmt.Println("Error creating HTTP OTLP exporter: ", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String("0.0.1"),
		attribute.String("environment", config.Environment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlpHTTPExporter),
	)
	otel.SetTracerProvider(tp)

	return tp
}

// StartSpan ...
func StartSpan(tracerCtx TraceContext, operation string, fn func(ctx context.Context, span trace.Span) error) error {
	tracer := tracerCtx.Tracer
	ctx := tracerCtx.RequestCtx
	ctx, span := tracer.Start(ctx, operation)
	defer span.End()

	return fn(ctx, span)
}
