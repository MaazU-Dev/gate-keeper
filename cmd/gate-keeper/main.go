package main

import (
	"context"
	"errors"
	"gate-keeper/internal/config"
	"gate-keeper/internal/proxy"
	"gate-keeper/internal/ratelimiter"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	godotenv.Load("../../.env")

	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		log.Fatalf("REDIS_ADDRESS is not set")
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatalf("SERVER_PORT is not set")
	}
	authTokenSecret := os.Getenv("AUTH_TOKEN_SECRET")
	if authTokenSecret == "" {
		log.Fatalf("AUTH_TOKEN_SECRET is not set")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}

	services, err := config.LoadServices(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	rl := ratelimiter.NewRateLimiter(redis.NewClient(&redis.Options{
		Addr: redisAddr,
	}))
	defer rl.Close()

	cfg := &config.Config{
		Services:        services,
		AuthTokenSecret: authTokenSecret,
	}

	mux := http.NewServeMux()
	proxy.RegisterRoutes(mux, cfg, rl)

	ongoingCtx, cancelFn := context.WithCancel(context.Background())
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ongoingCtx
		},
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on port %s", port)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	cancelFn()
	log.Println("server shutdown complete")
}
