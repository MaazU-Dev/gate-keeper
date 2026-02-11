package main

import (
	"fmt"
	"net/http"
)

func (cfg *Config) registerService(mux *http.ServeMux) {

	for _, service := range cfg.Services {
		for _, endpoint := range service.Endpoints {
			mux.HandleFunc(fmt.Sprintf("%s %s%s", endpoint.Method, service.BaseURL, endpoint.Path), func(w http.ResponseWriter, r *http.Request) {
				// service.Handler(w, r)
			})
		}
	}
}
