package main

import (
	"encoding/json"
	"fmt"
	ratelimiter "gate-keeper/internal/rate_limiter"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Services        []Service `json:"services"`
	AuthTokenSecret string    `json:"auth_token_secret"`
	RateLimiter     *ratelimiter.RateLimiter
}

type Service struct {
	Name        string          `json:"name"`
	BaseURL     string          `json:"base_url"`
	Port        int             `json:"port"`
	Endpoints   []Endpoint      `json:"endpoints"`
	SecretKey   string          `json:"secret_key"`
	IPFilter    IPFilter        `json:"ip_filter"`
	RateLimiter map[string]Rule `json:"rate_limiter"`
}

type Rule struct {
	Key   string `json:"key"`
	Rate  int    `json:"rate"`
	Burst int    `json:"burst"`
}

type RateLimitKey string

const (
	RateLimitKeyGlobal RateLimitKey = "global"
	RateLimitKeyIP     RateLimitKey = "ip"
	RateLimitKeyUser   RateLimitKey = "user"
	RateLimitKeyKey    RateLimitKey = "key"
)

type IPFilter struct {
	Mode string   `json:"mode"`
	IPs  []string `json:"ips"`
}
type Endpoint struct {
	Path         string       `json:"path"`
	Method       string       `json:"method"`
	AuthStrategy AuthStrategy `json:"auth_strategy"`
}

type AuthStrategy string

const (
	AuthStrategyJWT    AuthStrategy = "jwt"
	AuthStrategyPublic AuthStrategy = "public"
)

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
	godotenv.Load()
	redisAddr := os.Getenv("REDDIS_ADDRESS")
	if redisAddr == "" {
		log.Fatalf("REDDIS_ADDRESS is not set")
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatalf("SERVER_PORT is not set")
	}
	authTokenSecret := os.Getenv("AUTH_TOKEN_SECRET")
	if authTokenSecret == "" {
		log.Fatalf("AUTH_TOKEN_SECRET is not set")
	}

	services, err := loadServces("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cfg := Config{
		Services: services,
		RateLimiter: ratelimiter.NewRateLimiter(redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})),
		AuthTokenSecret: authTokenSecret,
	}

	mux := http.NewServeMux()
	cfg.registerService(mux)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
