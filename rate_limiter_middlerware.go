package main

import (
	"fmt"
	"log"
	"net/http"
)

func (cfg *Config) RateLimiterMiddleware(next http.Handler, service *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Client Identifiers
		userID, ok := ctx.Value("userID").(string)
		if !ok {
			userID = ""
		}
		apiKey := r.Header.Get("X-API-Key")
		ip := getClientIP(r)

		// {Key, Rate (per sec), Burst}
		rules := []struct {
			key   string
			rate  int
			burst int
		}{}

		// only add rules if they are configured for the service.
		if rl, ok := service.RateLimiter[string(RateLimitKeyGlobal)]; ok && rl.Key != "" {
			rules = append(rules, struct {
				key   string
				rate  int
				burst int
			}{
				rl.Key, rl.Rate, rl.Burst, // global Safety
			})
		}

		if rl, ok := service.RateLimiter[string(RateLimitKeyIP)]; ok && rl.Key != "" {
			rules = append(rules, struct {
				key   string
				rate  int
				burst int
			}{
				rl.Key + ":" + ip, rl.Rate, rl.Burst, // per IP
			})
		}

		if userID != "" {
			if rl, ok := service.RateLimiter[string(RateLimitKeyUser)]; ok && rl.Key != "" {
				fmt.Println("userID", userID)
				rules = append(rules, struct {
					key   string
					rate  int
					burst int
				}{
					rl.Key + ":" + userID, rl.Rate, rl.Burst, // per user (authenticated)
				})
			}
		}

		if apiKey != "" {
			if rl, ok := service.RateLimiter[string(RateLimitKeyKey)]; ok && rl.Key != "" {
				fmt.Println("apiKey", apiKey)
				rules = append(rules, struct {
					key   string
					rate  int
					burst int
				}{
					rl.Key + ":" + apiKey, rl.Rate, rl.Burst, // per API key
				})
			}
		}

		fmt.Println("Rules", rules)

		// execute checks (most restrictive logic)
		for _, rule := range rules {
			fmt.Println("Checking rate limit for", rule.key, rule.rate, rule.burst)
			res, err := cfg.RateLimiter.Check(ctx, rule.key, rule.rate, rule.burst)
			if err != nil {
				log.Println("Error checking rate limit", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if res[0] == 0 {
				w.Header().Set("X-RateLimit-Scope", rule.key)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			// if allowed == 1, but remaining is 0, means redis has daddy issues, so no head :/
			if res[1] != 0 {
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res[1]))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rule.burst))
			}
		}

		fmt.Println("Rate limit check passed")

		next.ServeHTTP(w, r)
	})
}
