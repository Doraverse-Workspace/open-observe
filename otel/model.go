package otel

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Config ...
// ServiceName: the name of the service
// Endpoint: the endpoint of the collector http or grpc. Example: localhost:4318 or localhost:4317
// IsSecure: whether the collector is secure true or false. If secure is true, the collector will use the https protocol.
// BasicAuth: the basic auth of the collector. Example: base64(admin:password)
type Config struct {
	ServiceName string
	Endpoint    string
	IsSecure    bool
	BasicAuth   string
	Environment string
	StreamName  string
}

type TraceContext struct {
	Tracer     trace.Tracer
	RequestCtx context.Context
}
