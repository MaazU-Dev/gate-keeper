package proxy

import (
	"fmt"
	"gate-keeper/internal/config"
	"gate-keeper/internal/middleware"
	"gate-keeper/internal/ratelimiter"
	"log"
	"net/http"
)

// middleware chain (logging -> IP filter -> auth -> rate limiter -> proxy).
func RegisterRoutes(mux *http.ServeMux, cfg *config.Config, rl *ratelimiter.RateLimiter) {
	for _, service := range cfg.Services {
		// one reverse proxy per service (reused across all its endpoints).
		p := NewServiceProxy(&service)

		for _, endpoint := range service.Endpoints {
			pattern := fmt.Sprintf("%s /%s%s", endpoint.Method, service.Name, endpoint.Path)
			log.Printf("registering %s  auth_strategy=%s\n", pattern, endpoint.AuthStrategy)

			baseHandler := Handler(p, endpoint)

			var finalHandler http.Handler = middleware.RateLimiterMiddleware(baseHandler, &service, rl)
			if endpoint.AuthStrategy == config.AuthStrategyJWT {
				finalHandler = middleware.AuthMiddleware(finalHandler, cfg.AuthTokenSecret)
			}
			finalHandler = middleware.IPFilterMiddleware(finalHandler, &service)
			finalHandler = middleware.LoggingMiddleware(finalHandler)

			mux.Handle(pattern, finalHandler)
		}
	}
}
