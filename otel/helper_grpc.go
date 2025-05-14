package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracerGRPC(config Config) *sdktrace.TracerProvider {
	gprcEndpoint := "127.0.0.1:5081"
	if config.Endpoint != "" {
		gprcEndpoint = config.Endpoint
	}

	streamName := "default"
	if config.StreamName != "" {
		streamName = config.StreamName
	}

	otlptracegrpc.NewClient()

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(gprcEndpoint),
		otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": "Basic " + config.BasicAuth,
			"stream-name":   streamName,
		}),
	}

	if !config.IsSecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	otlpGRPCExporter, err := otlptracegrpc.New(context.TODO(), opts...)

	if err != nil {
		fmt.Println("Error creating HTTP OTLP exporter: ", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String(config.ServiceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlpGRPCExporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
