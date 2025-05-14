# open-observe

# Install
```
go get -u github.com/Doraverse-Workspace/open-observe@latest
```

# Project Information

This document provides an overview of the project, instructions for integrating new modules, and details about the OpenTelemetry (otel) module.

## Integrating a New Module

Integrating a new module into this project generally involves the following steps. Please adapt these steps based on the specific nature of the module you are developing.

1.  **Define the Module's Purpose and Scope:**
    *   Clearly outline what the new module will do.
    *   Define its boundaries and interactions with other parts of the application.

2.  **Directory Structure:**
    *   Create a new directory for your module (e.g., `/<your_module_name>/`).
    *   Organize sub-directories for models, handlers, services, utilities, etc., as needed.

3.  **Core Logic Implementation:**
    *   Implement the primary functionality of your module.
    *   Write clean, maintainable, and well-documented code.

4.  **Configuration:**
    *   If your module requires configuration, decide how it will be managed (e.g., environment variables, configuration files).
    *   Provide clear instructions on how to configure the module.

5.  **API Endpoints (if applicable):**
    *   If the module exposes an API, define the routes and handlers.
    *   Ensure API endpoints are registered with the main application router.
    *   For web applications, consider using a framework like Echo and follow its conventions for defining routes and handlers.

6.  **Dependencies:**
    *   Add any new external dependencies to the `go.mod` file.
    *   Run `go mod tidy` to update `go.sum`.

7.  **Error Handling:**
    *   Implement robust error handling within your module.
    *   Ensure errors are propagated appropriately or handled gracefully.

8.  **Logging and Telemetry:**
    *   Integrate with the existing logging mechanism.
    *   If relevant, integrate with the OpenTelemetry setup for tracing and metrics. (See the `otel/` module documentation below).

9.  **Testing:**
    *   Write unit tests for the core logic of your module.
    *   Write integration tests to ensure your module works correctly with other parts of the application.

10. **Documentation:**
    *   Add comments to your code, especially for public functions and complex logic.
    *   Update this README or create a module-specific README to document its functionality, setup, and usage.

11. **Registration/Initialization:**
    *   Ensure your module is initialized and/or registered with the main application during startup. This might involve calling an initialization function from your module in the `main.go` or a relevant application setup file.

12. **Review and Merge:**
    *   Follow the project's contribution guidelines for code reviews and merging your changes.

## OpenTelemetry (otel) Module (`otel/`)

The `otel/` directory contains the OpenTelemetry integration for distributed tracing.

### File Descriptions

*   **`constants.go`**
    *   **Purpose:** Defines constant values used within the `otel` module.
    *   **Details:** Currently, it primarily defines `TraceContextKey`, which is used as a key for storing trace context in request contexts (e.g., in Echo framework).

*   **`helper_grpc.go`**
    *   **Purpose:** Provides a helper function to initialize the OpenTelemetry tracer provider with a gRPC OTLP (OpenTelemetry Protocol) exporter.
    *   **Details:**
        *   `InitTracerGRPC(config Config)`: This function sets up and configures the tracer provider to send trace data to an OTLP collector via gRPC.
        *   It takes a `Config` struct which includes service name, endpoint, security settings, and basic authentication credentials.
        *   It configures the gRPC exporter with the specified endpoint and headers (including authorization and stream name).
        *   It sets the global tracer provider and text map propagator for OpenTelemetry.

*   **`helper_http.go`**
    *   **Purpose:** Provides a helper function to initialize the OpenTelemetry tracer provider with an HTTP OTLP exporter and a utility function for starting spans.
    *   **Details:**
        *   `InitTracerHTTP(config Config)`: Similar to `InitTracerGRPC`, this function sets up the tracer provider to send trace data via HTTP to an OTLP collector. It uses a `Config` struct for settings like service name, endpoint, security, basic authentication, and stream name. It configures the HTTP exporter with the endpoint, URL path, and headers.
        *   `StartSpan(tracerCtx TraceContext, operation string, fn func(ctx context.Context, span trace.Span) error) error`: A utility function to simplify the creation and management of new spans. It takes a `TraceContext` (containing the tracer and request context), an operation name, and a function to execute within the span. The span is automatically ended when the function completes.

*   **`middleware.go`**
    *   **Purpose:** Provides Echo middleware for OpenTelemetry tracing.
    *   **Details:**
        *   `OtelMiddleware(config Config) echo.MiddlewareFunc`: This function returns an Echo middleware that automatically traces incoming HTTP requests.
        *   It extracts trace context from incoming request headers.
        *   It creates a new span for each request, naming it after the route or HTTP method.
        *   It enriches the span with standard HTTP attributes (method, target, route, host, scheme, client IP, user agent, request content length, request ID) and custom attributes from `TraceData`.
        *   It injects the trace context (tracer and request context with the active span) into the Echo context for use by downstream handlers.
        *   It records errors and updates span status if an error occurs during request processing.
        *   It adds response attributes like content length and content type.

*   **`model.go`**
    *   **Purpose:** Defines data structures (models) used within the `otel` module.
    *   **Details:**
        *   `Config`: A struct to hold configuration parameters for the OpenTelemetry setup. This includes `ServiceName`, `Endpoint` (for the OTLP collector), `IsSecure` (to use HTTPS/GRPCS), `BasicAuth` (for collector authentication), `Environment`, and `StreamName`.
        *   `TraceContext`: A struct to bundle an OpenTelemetry `trace.Tracer` and a `context.Context` together, typically for passing around tracing capabilities within the application.

*   **`trace_data.go`**
    *   **Purpose:** Defines a structure for custom trace data and provides functions to add this data as attributes to spans.
    *   **Details:**
        *   `TraceData`: A struct to hold various custom attributes that can be added to a span, such as `UserID`, `RequestID`, `ServiceName`, `Environment`, `Version`, `Action`, `Resource`, `StatusCode`, `Error`, `ClientIP`, `UserAgent`, request/response sizes, duration, and start/end times.
        *   `AddTraceAttributes(span trace.Span, data TraceData)`: A function that takes an active `trace.Span` and a `TraceData` object, then sets the fields from `TraceData` as attributes on the span. This is useful for enriching traces with application-specific information.
        *   `NewTraceData() TraceData`: A constructor function that creates a `TraceData` instance with some default values (e.g., environment, version, service name, region, start time).

---

## Code Examples

This section provides examples of how to use the `otel` module in your application.

### 1. Initializing the Tracer

You need to initialize a tracer provider (either HTTP or gRPC) at the beginning of your application.

**Using HTTP Exporter:**

```go
package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"your_project_path/otel" // Replace with your actual project path
)

func main() {
	// Configure the OTel tracer
	otelConfig := otel.Config{
		ServiceName: "my-sample-app",
		Endpoint:    "localhost:4318", // OTLP HTTP collector endpoint
		IsSecure:    false,
		BasicAuth:   "", // Optional: "dXNlcjpwYXNzd29yZA==" (base64_encode("user:password"))
		Environment: "development",
		StreamName:  "my-stream",
	}
	tp := otel.InitTracerHTTP(otelConfig)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	e := echo.New()

	// Add OtelMiddleware
	e.Use(otel.OtelMiddleware(otelConfig))

	// ... your routes and other setup ...
	// e.GET("/hello", helloHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

// func helloHandler(c echo.Context) error {
// 	 // Example of getting trace context and starting a new span
// 	 traceCtx, ok := c.Get(otel.TraceContextKey).(otel.TraceContext)
// 	 if !ok {
// 	 	return c.String(http.StatusInternalServerError, "could not get trace context")
// 	 }
//
// 	 err := otel.StartSpan(traceCtx, "helloHandlerLogic", func(ctx context.Context, span trace.Span) error {
// 	 	// Add custom attributes
// 	 	customData := otel.NewTraceData()
// 	 	customData.UserID = "user-123"
// 	 	customData.Action = "process_data"
// 	 	otel.AddTraceAttributes(span, customData)
//
// 	 	// Your handler logic here
// 	 	log.Println("Executing helloHandlerLogic")
// 	 	return nil
// 	 })
//
// 	 if err != nil {
// 	 	return c.String(http.StatusInternalServerError, err.Error())
// 	 }
// 	 return c.String(http.StatusOK, "Hello, World!")
// }

```

**Using gRPC Exporter:**

```go
package main

import (
	"log"
	"os"
	"context"

	"your_project_path/otel" // Replace with your actual project path
)

func main() {
	// Configure the OTel tracer
	otelConfig := otel.Config{
		ServiceName: "my-sample-app-grpc",
		Endpoint:    "localhost:4317", // OTLP gRPC collector endpoint
		IsSecure:    false,
		BasicAuth:   "", // Optional
		Environment: "staging",
		StreamName:  "my-grpc-stream",
	}
	tp := otel.InitTracerGRPC(otelConfig)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// ... rest of your application setup ...
	log.Println("Tracer initialized with gRPC exporter.")
	// Example: Start a gRPC server or client that uses this tracer
}
```

### 2. Using OtelMiddleware with Echo

The `OtelMiddleware` automatically traces incoming HTTP requests when using the Echo framework.

```go
// (Continuing from the HTTP Exporter example above)

// ...
// tp := otel.InitTracerHTTP(otelConfig)
// ...

e := echo.New()

// Add OtelMiddleware globally or to specific routes/groups
e.Use(otel.OtelMiddleware(otelConfig))

e.GET("/ping", func(c echo.Context) error {
	// This request will be automatically traced
	return c.String(http.StatusOK, "pong")
})

// ...
// e.Logger.Fatal(e.Start(":1323"))
// ...
```

### 3. Manually Starting a Span

You can manually create new spans to trace specific operations within your handlers or services.

```go
import (
	"context"
	"net/http"
	"log"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"your_project_path/otel" // Replace with your actual project path
)

func myCustomHandler(c echo.Context) error {
	// Retrieve the TraceContext set by the middleware
	traceCtxInterface := c.Get(otel.TraceContextKey)
	if traceCtxInterface == nil {
		log.Println("TraceContextKey not found in Echo context")
		return c.String(http.StatusInternalServerError, "Trace context not available")
	}

	traceCtx, ok := traceCtxInterface.(otel.TraceContext)
	if !ok {
		log.Println("Retrieved context is not of type otel.TraceContext")
		return c.String(http.StatusInternalServerError, "Invalid trace context type")
	}

	// Start a new span for a specific operation
	operationName := "complexDatabaseQuery"
	err := otel.StartSpan(traceCtx, operationName, func(ctx context.Context, span trace.Span) error {
		// Your logic for the operation goes here
		log.Printf("Executing operation: %s", operationName)
		// Simulate some work
		// time.Sleep(50 * time.Millisecond)

		// You can add attributes to this span too
		span.SetAttributes(attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"))

		// If an error occurs within the span
		// return errors.New("database query failed")
		return nil
	})

	if err != nil {
		// The error will be recorded on the span automatically by StartSpan if it bubbles up
		// Or handle it here
		return c.String(http.StatusInternalServerError, "Operation failed: "+err.Error())
	}

	return c.String(http.StatusOK, "Custom operation completed successfully")
}
```

### 4. Adding Custom Trace Attributes with `TraceData`

Enrich your spans with custom attributes for more detailed tracing.

```go
// (Inside a function where you have access to a trace.Span, e.g., within OtelMiddleware or StartSpan callback)

import (
	"go.opentelemetry.io/otel/trace"
	"your_project_path/otel" // Replace with your actual project path
)

func processRequestWithCustomData(span trace.Span, requestData map[string]interface{}) {
	traceData := otel.NewTraceData() // Initializes with some defaults

	// Populate TraceData with relevant information
	if userID, ok := requestData["userID"].(string); ok {
		traceData.UserID = userID
	}
	traceData.Action = "process_payment"
	traceData.Resource = "/api/payments"
	traceData.RequestID = "req-abc-123" // Or get from actual request
	// ... set other fields like Environment, Version, custom metrics etc.

	// Add these attributes to the current span
	otel.AddTraceAttributes(span, traceData)

	// ... rest of your processing logic ...
}

// Example usage within an Echo handler after getting the span:
// func someHandler(c echo.Context) error {
//    traceCtx := c.Get(otel.TraceContextKey).(otel.TraceContext)
//    _, span := traceCtx.Tracer.Start(traceCtx.RequestCtx, "someHandlerSpan")
//    defer span.End()
//
//    requestData := map[string]interface{}{"userID": "user-456"}
//    processRequestWithCustomData(span, requestData)
//
//    return c.String(http.StatusOK, "Processed with custom data")
// }
```
