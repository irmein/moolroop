package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/moolroop/docs"
	v1 "github.com/moolroop/gen/v1"
	"github.com/moolroop/internal/handler"
	"github.com/moolroop/internal/middleware"
	"github.com/moolroop/internal/store"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

// @title           MoolRoop Identity Core API
// @version         1.0
// @description     Minimalist Identity & Immutable Activity Engine
// @BasePath        /api/v1
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	memStore := store.NewMemoryStore()

	// 1. Launch gRPC Engine with Structured Slog Interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCUnaryLoggingInterceptor(logger)),
	)
	grpcHandler := handler.NewGrpcHandler(memStore)
	v1.RegisterIdentityServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("Failed to listen for gRPC", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// 2. Launch HTTP/REST & Swagger Engine with JSON slog Middleware
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger(logger))

	profileHandler := handler.NewProfileHandler(memStore)
	activityHandler := handler.NewActivityHandler(memStore)

	v1Group := r.Group("/api/v1")
	{
		v1Group.POST("/profiles", profileHandler.CreateProfile)
		v1Group.GET("/profiles/:id", profileHandler.GetProfile)
		v1Group.PATCH("/profiles/:id", profileHandler.PatchProfile)

		v1Group.POST("/activities", activityHandler.LogActivity)
		v1Group.GET("/activities/:user_id", activityHandler.ListActivities)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		slog.Info("HTTP & Swagger UI running on :8080 (UI at /swagger/index.html)")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// Graceful shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("Shutting down servers gracefully...", "signal", sig.String())

	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server forced to shutdown", "error", err)
	}

	slog.Info("Servers exited successfully")
}
