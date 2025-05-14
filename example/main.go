package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Doraverse-Workspace/open-observe/example/handler"
	db "github.com/Doraverse-Workspace/open-observe/example/mongo"
	tel "github.com/Doraverse-Workspace/open-observe/otel"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config := tel.Config{
		ServiceName: "echo-v4-server-demo-tracer",
		Endpoint:    "....", // TODO: Replace with your endpoint
		IsSecure:    true,
		BasicAuth:   "....", // TODO: Replace with your basic auth credentials
		Environment: "development",
		StreamName:  "default",
	}
	// Initialize OpenTelemetry
	tp := tel.InitTracerHTTP(config)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Println("Error shutting down tracer provider: ", err)
		}
	}()

	// Initialize MongoDB client with APM configuration
	mongoConfig := db.DefaultConfig()
	mongoConfig.URI = "mongodb://localhost:27017"
	mongoConfig.Database = "userdb"
	mongoConfig.APMConfig = db.APMConfig{
		SlowOperationThreshold: 100 * time.Millisecond,
		EnableCommandMonitor:   true,
		EnablePoolMonitor:      true,
	}
	mongoConfig.MinPoolSize = 5
	mongoConfig.MaxPoolSize = 100
	mongoConfig.MaxConnIdleTime = 5 * time.Minute
	mongoConfig.Timeout = 10 * time.Second

	mongoClient, err := db.NewClient(context.Background(), mongoConfig)
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v\n", err)
		return
	}
	defer mongoClient.Close(context.Background())

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(tel.OtelMiddleware(config))

	// Create user handler
	userHandler := handler.NewUserHandler(mongoClient)

	// User routes
	e.POST("/users", userHandler.CreateUser)

	e.GET("/users", userHandler.ListUsers)

	// Health check
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}
