package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Services []Service `json:"services"`
}

type Service struct {
	Name      string     `json:"name"`
	BaseURL   string     `json:"base_url"`
	Port      int        `json:"port"`
	Endpoints []Endpoint `json:"endpoints"`
	SecretKey string     `json:"secret_key"`
}
type Endpoint struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func loadServces(path string) ([]Service, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var services []Service
	if err := json.NewDecoder(f).Decode(&services); err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no services found in config")
	}

	return services, nil
}

func main() {
	services, err := loadServces("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cfg := Config{
		Services: services,
	}

	mux := http.NewServeMux()
	cfg.registerService(mux)
	port := "8080"
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
