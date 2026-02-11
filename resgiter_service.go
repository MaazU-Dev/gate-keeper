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

			fmt.Printf("%s /%s%s \n", endpoint.Method, service.Name, endpoint.Path)
			mux.HandleFunc(fmt.Sprintf("%s /%s%s", endpoint.Method, service.Name, endpoint.Path), func(w http.ResponseWriter, r *http.Request) {
				log.Println("request received", r.Method, r.URL.Path)
				service.proxyHandler(w, r, endpoint)
			})
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
	proxy.ServeHTTP(w, r)
}
