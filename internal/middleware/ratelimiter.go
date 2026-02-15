package middleware

import (
	"fmt"
	"gate-keeper/internal/config"
	"gate-keeper/internal/ratelimiter"
	"log"
	"net/http"
)

func RateLimiterMiddleware(next http.Handler, service *config.Service, rl *ratelimiter.RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Client identifiers
		userID, _ := ctx.Value(config.UserIDKey).(string)
		apiKey := r.Header.Get("X-API-Key")
		ip := GetClientIP(r)

		type rule struct {
			key   string
			rate  int
			burst int
		}
		var rules []rule

		if v, ok := service.RateLimiter[string(config.RateLimitKeyGlobal)]; ok && v.Key != "" {
			rules = append(rules, rule{v.Key, v.Rate, v.Burst})
		}

		if v, ok := service.RateLimiter[string(config.RateLimitKeyIP)]; ok && v.Key != "" {
			rules = append(rules, rule{v.Key + ":" + ip, v.Rate, v.Burst})
		}

		if userID != "" {
			if v, ok := service.RateLimiter[string(config.RateLimitKeyUser)]; ok && v.Key != "" {
				rules = append(rules, rule{v.Key + ":" + userID, v.Rate, v.Burst})
			}
		}

		if apiKey != "" {
			if v, ok := service.RateLimiter[string(config.RateLimitKeyKey)]; ok && v.Key != "" {
				rules = append(rules, rule{v.Key + ":" + apiKey, v.Rate, v.Burst})
			}
		}

		// Execute checks (most restrictive first)
		for _, r := range rules {
			res, err := rl.Check(ctx, r.key, r.rate, r.burst)
			if err != nil {
				log.Println("error checking rate limit:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if res[0] == 0 {
				w.Header().Set("X-RateLimit-Scope", r.key)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			if res[1] != 0 {
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res[1]))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.burst))
			}
		}

		next.ServeHTTP(w, r)
	})
}
