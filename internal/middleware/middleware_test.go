package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moolroop/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRequestIDMiddleware_Generated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())

	var capturedID string
	r.GET("/test", func(c *gin.Context) {
		capturedID = middleware.GetRequestID(c)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	headerID := w.Header().Get(middleware.RequestIDHeader)
	assert.NotEmpty(t, headerID)
	assert.Equal(t, headerID, capturedID)
}

func TestRequestIDMiddleware_Supplied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())

	customID := "custom-req-12345"
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(middleware.RequestIDHeader, customID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, customID, w.Header().Get(middleware.RequestIDHeader))
}

func TestStructuredLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger(logger))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var logEntry map[string]any
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "HTTP request processed", logEntry["msg"])
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/ping", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.NotEmpty(t, logEntry["request_id"])
}

func TestGRPCUnaryLoggingInterceptor(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	interceptor := middleware.GRPCUnaryLoggingInterceptor(logger)

	// Test Success
	handlerSuccess := func(ctx context.Context, req any) (any, error) {
		return "response-ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/moolroop.v1.IdentityService/GetProfile"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "grpc-trace-999"))

	resp, err := interceptor(ctx, "req", info, handlerSuccess)
	require.NoError(t, err)
	assert.Equal(t, "response-ok", resp)

	var logEntry map[string]any
	err = json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "gRPC request completed", logEntry["msg"])
	assert.Equal(t, "grpc-trace-999", logEntry["request_id"])
	assert.Equal(t, "/moolroop.v1.IdentityService/GetProfile", logEntry["grpc_method"])
	assert.Equal(t, "OK", logEntry["grpc_code"])

	// Test Error
	logBuf.Reset()
	handlerErr := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "user profile not found")
	}

	resp, err = interceptor(context.Background(), "req", info, handlerErr)
	require.Error(t, err)
	assert.Nil(t, resp)

	var logErrEntry map[string]any
	err = json.Unmarshal(logBuf.Bytes(), &logErrEntry)
	require.NoError(t, err)
	assert.Equal(t, "gRPC request completed with error", logErrEntry["msg"])
	assert.Equal(t, "NotFound", logErrEntry["grpc_code"])
}
