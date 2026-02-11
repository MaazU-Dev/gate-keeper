package main

import (
	"fmt"
	"net/http"
)

func (cfg *Config) resolveRequest(r *http.Request) (*Service, error) {
	return nil, fmt.Errorf("no matching service for %s %s", r.Method, r.URL.Path)
}
