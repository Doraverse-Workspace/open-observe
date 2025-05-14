package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Config represents MongoDB configuration
type Config struct {
	URI             string
	Database        string
	Username        string
	Password        string
	MinPoolSize     uint64
	MaxPoolSize     uint64
	MaxConnIdleTime time.Duration
	Timeout         time.Duration
	APMConfig       APMConfig
}

// APMConfig represents Application Performance Monitoring configuration
type APMConfig struct {
	SlowOperationThreshold time.Duration // Threshold for slow operation logging
	EnableCommandMonitor   bool          // Enable command monitoring
	EnablePoolMonitor      bool          // Enable connection pool monitoring
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		URI:             "mongodb://localhost:27017",
		Database:        "test",
		MinPoolSize:     5,
		MaxPoolSize:     100,
		MaxConnIdleTime: 5 * time.Minute,
		Timeout:         10 * time.Second,
		APMConfig: APMConfig{
			SlowOperationThreshold: 100 * time.Millisecond,
			EnableCommandMonitor:   true,
			EnablePoolMonitor:      true,
		},
	}
}

// Client wraps the MongoDB client with tracing capabilities
type Client struct {
	client   *mongo.Client
	database string
	tracer   trace.Tracer
}

// NewClient creates a new MongoDB client with tracing and monitoring
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// Create MongoDB client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	if cfg.Username != "" && cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	// Configure command monitoring if enabled
	if cfg.APMConfig.EnableCommandMonitor {
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
				span := trace.SpanFromContext(ctx)
				span.SetAttributes(
					attribute.String("db.operation", evt.CommandName),
					attribute.String("db.statement", evt.Command.String()),
					attribute.String("db.connection_id", evt.ConnectionID),
				)
			},
			Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
				span := trace.SpanFromContext(ctx)
				span.SetAttributes(
					attribute.Int64("db.duration_ms", evt.DurationNanos/1e6),
				)
			},
			Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
				span := trace.SpanFromContext(ctx)
				span.SetStatus(codes.Error, evt.Failure)
				span.RecordError(fmt.Errorf("mongodb command failed: %s", evt.Failure))
			},
		}
		clientOptions.SetMonitor(monitor)
	}

	// Configure connection pool monitoring if enabled
	if cfg.APMConfig.EnablePoolMonitor {
		poolMonitor := &event.PoolMonitor{
			Event: func(evt *event.PoolEvent) {
				// Log pool events for monitoring
				fmt.Printf("MongoDB Pool Event: Type=%s, Address=%s, ConnectionID=%d\n",
					evt.Type, evt.Address, evt.ConnectionID)
			},
		}
		clientOptions.SetPoolMonitor(poolMonitor)
	}

	// Connect to MongoDB with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	client, err := mongo.Connect(timeoutCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(timeoutCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: cfg.Database,
		tracer:   otel.Tracer("mongodb"),
	}, nil
}

// Collection returns a MongoDB collection with tracing wrapper
func (c *Client) Collection(name string) *Collection {
	return &Collection{
		coll:   c.client.Database(c.database).Collection(name),
		tracer: c.tracer,
	}
}

// Close disconnects from MongoDB
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// startSpan starts a new span for MongoDB operation
func (c *Collection) startSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	return c.tracer.Start(ctx, fmt.Sprintf("MongoDB.%s", operation),
		trace.WithAttributes(
			attribute.String("db.system", "mongodb"),
			attribute.String("db.name", c.coll.Database().Name()),
			attribute.String("db.collection", c.coll.Name()),
			attribute.String("db.operation", operation),
		),
	)
}

// handleError handles error and sets span status
func handleError(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// Collection wraps MongoDB collection with tracing
type Collection struct {
	coll   *mongo.Collection
	tracer trace.Tracer
}

// InsertOne inserts a document with automatic span creation
func (c *Collection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	ctx, span := c.startSpan(ctx, "InsertOne")
	defer span.End()

	result, err := c.coll.InsertOne(ctx, document, opts...)
	handleError(span, err)
	return result, err
}

// FindOne finds a single document with automatic span creation
func (c *Collection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	ctx, span := c.startSpan(ctx, "FindOne")
	defer span.End()

	result := c.coll.FindOne(ctx, filter, opts...)
	handleError(span, result.Err())
	return result
}

// Find finds multiple documents with automatic span creation
func (c *Collection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	ctx, span := c.startSpan(ctx, "Find")
	defer span.End()

	cursor, err := c.coll.Find(ctx, filter, opts...)
	handleError(span, err)
	return cursor, err
}

// UpdateOne updates a single document with automatic span creation
func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx, span := c.startSpan(ctx, "UpdateOne")
	defer span.End()

	result, err := c.coll.UpdateOne(ctx, filter, update, opts...)
	handleError(span, err)
	return result, err
}

// DeleteOne deletes a single document with automatic span creation
func (c *Collection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	ctx, span := c.startSpan(ctx, "DeleteOne")
	defer span.End()

	result, err := c.coll.DeleteOne(ctx, filter, opts...)
	handleError(span, err)
	return result, err
}
