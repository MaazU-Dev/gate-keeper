package middleware

import (
	"gate-keeper/internal/config"
	"log"
	"net"
	"net/http"
	"slices"
	"strings"
)

func ipInList(ip string, list []string) bool {
	if slices.Contains(list, ip) {
		return true
	}
	parsed := net.ParseIP(ip)
	if parsed != nil && parsed.IsLoopback() {
		return true
	}
	return false
}

// preferring X-Forwarded-For when behind a reverse proxy or load balancer.
func GetClientIP(r *http.Request) string {
	var ip string

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip = strings.TrimSpace(parts[0])
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		} else {
			ip = host
		}
	}

	return ip
}

// enforces whitelist / blacklist IP filtering for a service.
func IPFilterMiddleware(next http.Handler, service *config.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		inList := ipInList(ip, service.IPFilter.IPs)
		switch service.IPFilter.Mode {
		case "whitelist":
			if !inList {
				log.Println("ip is not whitelisted")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		case "blacklist":
			if inList {
				log.Println("ip is blacklisted")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		default:
			log.Println("No IP filter mode specified")
		}
		next.ServeHTTP(w, r)
	})
}
