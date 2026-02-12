package main

import (
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
		log.Println("invalid ip address")
		return true
	}
	return false
}

func getClientIP(r *http.Request) string {
	var ip string

	// Prefer the first IP in X-Forwarded-For when present (behind a proxy / load balancer)
	// When using Cloudfront the client IP address is in the x-original-forwarded-for (need to confirm this)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		log.Println("request is proxied through a load balancer")
		parts := strings.Split(xff, ",")
		ip = strings.TrimSpace(parts[0])
	} else {
		log.Println("request is not proxied through a load balancer")
		// r.RemoteAddr is usually "ip:port" strip the port so config
		// can just specify the raw IP (e.g. 127.0.0.1 or ::1).
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		} else {
			ip = host
		}
	}

	return ip
}

func IPFilterMiddleware(next http.Handler, service *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
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
