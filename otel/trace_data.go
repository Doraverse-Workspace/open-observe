package otel

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceData represents the structure for additional tracing information
type TraceData struct {
	UserID       string
	RequestID    string
	ServiceName  string
	Environment  string
	Version      string
	Region       string
	Action       string
	Resource     string
	StatusCode   int
	Error        error
	ClientIP     string    `json:"client_ip,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	RequestSize  int64     `json:"request_size,omitempty"`
	ResponseSize int64     `json:"response_size,omitempty"`
	Duration     float64   `json:"duration,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	EndTime      time.Time `json:"end_time,omitempty"`
}

// AddTraceAttributes adds custom attributes to a span
func AddTraceAttributes(span trace.Span, data TraceData) {
	// Add custom attributes to the span
	attributes := []attribute.KeyValue{
		attribute.String("user.id", data.UserID),
		attribute.String("request.id", data.RequestID),
		attribute.String("service.name", data.ServiceName),
		attribute.String("environment", data.Environment),
		attribute.String("version", data.Version),
		attribute.String("region", data.Region),
		attribute.String("action", data.Action),
		attribute.String("resource", data.Resource),
		attribute.Int("status.code", data.StatusCode),
		attribute.String("client.ip", data.ClientIP),
		attribute.String("user.agent", data.UserAgent),
		attribute.Int64("request.size", data.RequestSize),
		attribute.Int64("response.size", data.ResponseSize),
		attribute.Float64("duration.ms", data.Duration),
	}

	if !data.StartTime.IsZero() {
		attributes = append(attributes, attribute.String("start_time", data.StartTime.Format(time.RFC3339)))
	}

	if !data.EndTime.IsZero() {
		attributes = append(attributes, attribute.String("end_time", data.EndTime.Format(time.RFC3339)))
	}

	if data.Error != nil {
		attributes = append(attributes, attribute.String("error.message", data.Error.Error()))
	}

	if data.StatusCode != 0 {
		if data.StatusCode == http.StatusOK {
			span.SetStatus(codes.Ok, "Success")
		} else {
			if data.Error != nil {
				span.SetStatus(codes.Error, data.Error.Error())
			} else {
				span.SetStatus(codes.Error, "Unknown error")
			}
		}
	}

	span.SetAttributes(attributes...)
}

// NewTraceData creates a new TraceData instance with default values
func NewTraceData() TraceData {
	return TraceData{
		Environment: "development", // Default environment
		Version:     "1.0.0",       // Default version
		ServiceName: "echo-server",
		Region:      "local",
		StartTime:   time.Now(),
	}
}

// NewTracer returns a new tracer for the given service name
func NewTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}

// TraceError traces an error and records it in the span
func TraceError(span trace.Span, err error) {
	AddTraceAttributes(span, TraceData{
		StatusCode: 400,
		Error:      err,
		EndTime:    time.Now(),
	})
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// TraceSuccess traces a success and records it in the span
func TraceSuccess(span trace.Span) {
	AddTraceAttributes(span, TraceData{
		StatusCode: 200,
		EndTime:    time.Now(),
	})
	span.SetStatus(codes.Ok, "Success")
}
