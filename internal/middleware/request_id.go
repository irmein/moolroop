package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	// RequestIDHeader is the canonical header used for correlation and request tracing.
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for the correlation request ID.
	RequestIDContextKey contextKey = "RequestID"
)

// RequestID injects an X-Request-ID into the request context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		c.Header(RequestIDHeader, reqID)
		c.Set(string(RequestIDContextKey), reqID)

		ctx := context.WithValue(c.Request.Context(), RequestIDContextKey, reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the Gin context or returns empty.
func GetRequestID(c *gin.Context) string {
	if val, ok := c.Get(string(RequestIDContextKey)); ok {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
