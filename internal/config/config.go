package config

type Config struct {
	Services        []Service
	AuthTokenSecret string
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

type CtxKey string

const (
	TraceIDKey CtxKey = "trace_id"
	UserIDKey  CtxKey = "user_id"
)
