package proxy

import (
	"fmt"
	"gate-keeper/internal/config"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewServiceProxy(service *config.Service) *httputil.ReverseProxy {
	target, err := url.Parse(fmt.Sprintf("%s:%d", service.BaseURL, service.Port))
	if err != nil {
		log.Fatalf("failed to parse target for service %s: %v", service.Name, err)
	}
	return httputil.NewSingleHostReverseProxy(target)
}

func Handler(p *httputil.ReverseProxy, endpoint config.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = endpoint.Path

		// Set headers before proxying — the default Director copies them.
		if endpoint.AuthStrategy == config.AuthStrategyJWT {
			r.Header.Set("X-User-ID", r.Context().Value(config.UserIDKey).(string))
		}
		r.Header.Set("X-Request-ID", r.Context().Value(config.TraceIDKey).(string))

		p.ServeHTTP(w, r)
	}
}
