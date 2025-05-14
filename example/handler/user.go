package handler

import (
	"context"
	"net/http"
	"time"

	db "github.com/Doraverse-Workspace/open-observe/example/mongo"
	"github.com/Doraverse-Workspace/open-observe/example/mongo/model"
	tracermodule "github.com/Doraverse-Workspace/open-observe/otel"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/otel/trace"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	collection *db.Collection
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(client *db.Client) *UserHandler {
	return &UserHandler{
		collection: client.Collection("users"),
	}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c echo.Context) error {
	var user model.User
	tracerCtx := c.Get(tracermodule.TraceContextKey).(tracermodule.TraceContext)
	err := tracermodule.StartSpan(tracerCtx, "CreateUser", func(ctx context.Context, span trace.Span) error {
		if err := c.Bind(&user); err != nil {
			span.RecordError(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Set timestamps
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		_, err := h.collection.InsertOne(ctx, user)
		if err != nil {
			span.RecordError(err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}

// ListUsers lists all users
func (h *UserHandler) ListUsers(c echo.Context) error {
	var users []model.User
	time.Sleep(1 * time.Second)
	cursor, err := h.collection.Find(c.Request().Context(), bson.M{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer cursor.Close(c.Request().Context())
	if err := cursor.All(c.Request().Context(), &users); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}
