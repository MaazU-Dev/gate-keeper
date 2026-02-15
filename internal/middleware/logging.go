package middleware

import (
	"context"
	"gate-keeper/internal/config"
	"gate-keeper/internal/httputil"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		traceID := r.Header.Get("X-Request-ID") //propagates a trace ID (X-Request-ID) through the context and later on, header
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), config.TraceIDKey, traceID)
		w.Header().Set("X-Request-ID", traceID)

		wrapped := httputil.NewResponseWriter(w)

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		latency := time.Since(start)

		logger.Info("request_completed",
			slog.String("trace_id", traceID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.StatusCode()),
			slog.String("ip", r.RemoteAddr),
			slog.Duration("latency", latency),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}
