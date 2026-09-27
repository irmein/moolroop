package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCUnaryLoggingInterceptor provides structured slog JSON logging and correlation ID propagation for gRPC.
func GRPCUnaryLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		var reqID string
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			vals := md.Get("x-request-id")
			if len(vals) > 0 && vals[0] != "" {
				reqID = vals[0]
			}
		}
		if reqID == "" {
			reqID = uuid.New().String()
		}

		_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", reqID))
		ctx = context.WithValue(ctx, RequestIDContextKey, reqID)

		resp, err := handler(ctx, req)
		latency := time.Since(start)

		st, _ := status.FromError(err)
		statusCode := st.Code()

		attrs := []slog.Attr{
			slog.String("request_id", reqID),
			slog.String("grpc_method", info.FullMethod),
			slog.String("grpc_code", statusCode.String()),
			slog.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			logger.LogAttrs(ctx, slog.LevelError, "gRPC request completed with error", attrs...)
		} else {
			logger.LogAttrs(ctx, slog.LevelInfo, "gRPC request completed", attrs...)
		}

		return resp, err
	}
}
