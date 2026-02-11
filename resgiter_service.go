package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (cfg *Config) registerService(mux *http.ServeMux) {
	for _, service := range cfg.Services {
		for _, endpoint := range service.Endpoints {

			pattern := fmt.Sprintf("%s /%s%s", endpoint.Method, service.Name, endpoint.Path)
			fmt.Printf("%s %s auth strategy: %s\n", endpoint.Method, pattern, endpoint.AuthStrategy)

			baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println("request received", r.Method, r.URL.Path)
				service.proxyHandler(w, r, endpoint)
			})

			var finalHandler http.Handler = baseHandler
			if endpoint.AuthStrategy == AuthStrategyJWT {
				finalHandler = cfg.AuthMiddleware(finalHandler)
			}
			finalHandler = IPFilterMiddleware(finalHandler, &service)
			mux.Handle(pattern, finalHandler)
		}
	}
}

func (service *Service) proxyHandler(w http.ResponseWriter, r *http.Request, endpoint Endpoint) {
	r.URL.Path = endpoint.Path

	target, err := url.Parse(fmt.Sprintf("%s:%d", service.BaseURL, service.Port)) // e.g. http://BASE_URL:PORT
	if err != nil {
		log.Fatalf("failed to parse target: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		if endpoint.AuthStrategy == AuthStrategyJWT {
			r.Header.Set("X-User-ID", r.Context().Value("userID").(string))
		}
	}
	proxy.ServeHTTP(w, r)
}
