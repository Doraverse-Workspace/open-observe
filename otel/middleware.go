package otel

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// OtelMiddleware returns a middleware that will trace incoming requests.
func OtelMiddleware(config Config) echo.MiddlewareFunc {
	if config.ServiceName == "" {
		config.ServiceName = "default"
	}
	tracer := otel.Tracer(config.ServiceName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			// Extract trace information from the incoming request
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

			path := c.Path()
			route := path
			if route == "" {
				route = fmt.Sprintf("HTTP %s route not found", req.Method)
			}

			// Generate request ID if not exists
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				c.Response().Header().Set("X-Request-ID", requestID)
			}

			// Create trace data
			traceData := NewTraceData()
			traceData.RequestID = requestID
			traceData.Action = req.Method
			traceData.Resource = route
			traceData.Environment = config.Environment

			// Enhanced attributes for better tracing
			opts := []trace.SpanStartOption{
				// Standard HTTP attributes
				trace.WithAttributes(semconv.HTTPMethodKey.String(req.Method)),
				trace.WithAttributes(semconv.HTTPTargetKey.String(req.URL.Path)),
				trace.WithAttributes(semconv.HTTPRouteKey.String(route)),
				trace.WithAttributes(semconv.HTTPHostKey.String(req.Host)),
				trace.WithAttributes(semconv.HTTPSchemeKey.String(req.URL.Scheme)),
				trace.WithAttributes(attribute.String("http.flavor", req.Proto)),

				// Additional context attributes
				trace.WithAttributes(attribute.String("http.client_ip", c.RealIP())),
				trace.WithAttributes(attribute.String("http.user_agent", req.UserAgent())),
				trace.WithAttributes(attribute.Int64("http.request_content_length", req.ContentLength)),
				trace.WithAttributes(attribute.String("http.request_id", requestID)),
			}

			spanName := route
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s", req.Method)
			}

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// Add trace data attributes
			AddTraceAttributes(span, traceData)

			// Pass the span through the request context
			c.SetRequest(req.WithContext(ctx))
			c.Set(TraceContextKey, TraceContext{
				Tracer:     tracer,
				RequestCtx: ctx,
			})

			err := next(c)
			if err != nil {
				traceData.Error = err
				traceData.StatusCode = c.Response().Status
				AddTraceAttributes(span, traceData)
				span.RecordError(err)
				return err
			}

			// Update status code and response size in trace data
			traceData.StatusCode = c.Response().Status
			AddTraceAttributes(span, traceData)

			// Add response attributes
			span.SetAttributes(
				attribute.Int("http.response_content_length", int(c.Response().Size)),
				attribute.String("http.response_content_type", c.Response().Header().Get(echo.HeaderContentType)),
			)

			return nil
		}
	}
}
