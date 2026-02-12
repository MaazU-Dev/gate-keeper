package main

import (
	"fmt"
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

		// Only add rules if they are configured for the service.
		if rl, ok := service.RateLimiter[string(RateLimitKeyGlobal)]; ok && rl.Key != "" {
			rules = append(rules, struct {
				key   string
				rate  int
				burst int
			}{
				rl.Key, rl.Rate, rl.Burst, // Global Safety
			})
		}

		if rl, ok := service.RateLimiter[string(RateLimitKeyIP)]; ok && rl.Key != "" {
			rules = append(rules, struct {
				key   string
				rate  int
				burst int
			}{
				rl.Key + ":" + ip, rl.Rate, rl.Burst, // Per IP
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
					rl.Key + ":" + userID, rl.Rate, rl.Burst, // Per User (Authenticated)
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
					rl.Key + ":" + apiKey, rl.Rate, rl.Burst, // Per API Key
				})
			}
		}

		fmt.Println("Rules", rules)

		// 3. Execute Checks (Most restrictive logic)
		for _, rule := range rules {
			fmt.Println("Checking rate limit for", rule.key, rule.rate, rule.burst)
			if !cfg.RateLimiter.Check(ctx, rule.key, rule.rate, rule.burst) {
				w.Header().Set("X-RateLimit-Scope", rule.key)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
		}

		fmt.Println("Rate limit check passed")

		next.ServeHTTP(w, r)
	})
}
