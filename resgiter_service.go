package main

import (
	"fmt"
	"log"
	"net/http"
)

func (cfg *Config) registerService(mux *http.ServeMux) {

	for _, service := range cfg.Services {
		for _, endpoint := range service.Endpoints {

			fmt.Printf("%s /%s%s \n", endpoint.Method, service.Name, endpoint.Path)
			mux.HandleFunc(fmt.Sprintf("%s /%s%s", endpoint.Method, service.Name, endpoint.Path), func(w http.ResponseWriter, r *http.Request) {
				log.Println("request received", r.Method, r.URL.Path)
				// service.Handler(w, r)
			})
		}
	}
}
