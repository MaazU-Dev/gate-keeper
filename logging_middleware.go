package main

import (
	"context"
	responsewriter "gate-keeper/internal/response_writer"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

const TraceIdKey ctxKey = "trace_Id"

func LoggingMiddleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		traceId := r.Header.Get("X-Request-ID")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), TraceIdKey, traceId)
		w.Header().Set("X-Request-ID", traceId)

		wrapped := responsewriter.NewResponseWriter(w)

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		latency := time.Since(start)

		logger.Info("request_completed",
			slog.String("trace_id", traceId),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.StatusCode()),
			slog.String("ip", r.RemoteAddr),
			slog.Duration("latency", latency),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}
