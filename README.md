# Gate Keeper

A lightweight, config-driven API gateway written in Go. Gate Keeper sits in front of your backend services and handles cross-cutting concerns — authentication, rate limiting, IP filtering, and request logging — so your services don't have to.

## How It Works

Gate Keeper reads a JSON config file that describes your backend services and their endpoints. At startup it registers a reverse proxy for each service, wraps every route in a middleware chain, and forwards allowed traffic to the appropriate backend.

```
Client Request
      │
      ▼
┌─────────────┐
│   Logging    │  ← Assigns trace ID, logs method/path/status/latency as JSON
├─────────────┤
│  IP Filter   │  ← Whitelist or blacklist mode per service
├─────────────┤
│   Auth       │  ← JWT validation (skipped for public endpoints)
├─────────────┤
│ Rate Limiter │  ← Multi-scope token bucket via Redis
├─────────────┤
│ Reverse Proxy│  ← Forwards to backend service
└─────────────┘
      │
      ▼
  Backend Service
```

Every request flows through this chain top-to-bottom. If any middleware rejects the request, processing stops and an appropriate HTTP error is returned immediately.

## Features

- **Config-Driven** — Define services, endpoints, auth strategies, IP rules, and rate limits in a single JSON file. No code changes needed to onboard a new service.
- **Reverse Proxy** — Transparently forwards requests to backend services. One proxy is created per service at startup and reused across all requests.
- **JWT Authentication** — Validates Bearer tokens on protected endpoints and propagates the user ID to backends via the `X-User-ID` header.
- **IP Filtering** — Per-service whitelist or blacklist mode. Supports `X-Forwarded-For` for deployments behind a load balancer.
- **Rate Limiting** — Redis-backed token bucket algorithm with four independent scopes: global, per-IP, per-user, and per-API-key. Fail-open design — if Redis goes down, traffic is allowed through.
- **Structured Logging** — Every request is logged as JSON via `slog` with trace ID, method, path, status code, latency, IP, and user agent.
- **Request Tracing** — Generates a UUID trace ID for each request (or respects an incoming `X-Request-ID` header) and propagates it to backend services.
- **Graceful Shutdown** — Handles `SIGINT` and `SIGTERM` with a 10-second drain period so in-flight requests can complete.

## Architecture

```
gate-keeper/
├── cmd/
│   └── gate-keeper/
│       └── main.go                  # Entrypoint — wires dependencies, starts server
├── internal/
│   ├── auth/
│   │   └── jwt.go                   # JWT parsing and Bearer token extraction
│   ├── config/
│   │   ├── config.go                # Type definitions (Service, Endpoint, Rule, etc.)
│   │   └── loader.go                # JSON config file loader
│   ├── httputil/
│   │   └── response_writer.go       # ResponseWriter wrapper that captures status codes
│   ├── middleware/
│   │   ├── auth.go                  # JWT authentication middleware
│   │   ├── ipfilter.go              # IP whitelist / blacklist middleware
│   │   ├── logging.go               # Structured JSON request logging middleware
│   │   └── ratelimiter.go           # Multi-scope rate limiting middleware
│   ├── proxy/
│   │   ├── proxy.go                 # Reverse proxy creation and request handler
│   │   └── register.go              # Route registration and middleware chain wiring
│   └── ratelimiter/
│       ├── ratelimiter.go           # Redis token bucket client
│       └── ratelimiter.lua          # Atomic Lua script executed inside Redis
├── configs/
│   ├── config.json                  # Service definitions (your actual config)
│   └── config.example.json          # Example config safe to commit
├── .env.example                     # Environment variable template
├── Dockerfile                       # Multi-stage production build
├── docker-compose.yml               # Local dev setup (app + Redis)
├── Makefile                         # Build, run, test, lint targets
├── go.mod
└── go.sum
```

All application code lives under `internal/` following the standard Go project layout, meaning nothing is importable by external modules.

## Deep Dive

### Config Format

Services are defined as a JSON array in `configs/config.json`. Each service describes its backend address, endpoints, auth strategy, IP filtering rules, and rate limiting buckets.

```json
[
  {
    "name": "my-service",
    "base_url": "http://localhost",
    "port": 8081,
    "ip_filter": {
      "mode": "whitelist",
      "ips": ["127.0.0.1"]
    },
    "rate_limiter": {
      "global": { "key": "limit:global", "rate": 50, "burst": 50 },
      "ip":     { "key": "limit:ip",     "rate": 10, "burst": 100 },
      "user":   { "key": "limit:user",   "rate": 5,  "burst": 10 },
      "key":    { "key": "limit:key",    "rate": 5,  "burst": 10 }
    },
    "endpoints": [
      { "path": "/api/v1/users", "method": "GET",  "auth_strategy": "jwt" },
      { "path": "/api/v1/users", "method": "POST", "auth_strategy": "public" }
    ],
    "secret_key": "your-service-secret"
  }
]
```

Routes are registered as `METHOD /service-name/path`, so the GET endpoint above becomes `GET /my-service/api/v1/users` on the gateway.

### Authentication

Endpoints with `"auth_strategy": "jwt"` require a valid Bearer token in the `Authorization` header. The gateway:

1. Extracts the token from the `Authorization: Bearer <token>` header.
2. Validates the JWT signature using the `AUTH_TOKEN_SECRET` environment variable.
3. Extracts the `user_id` claim from the token payload.
4. Injects the user ID into the request context and forwards it to the backend as an `X-User-ID` header.

Endpoints with `"auth_strategy": "public"` skip authentication entirely.

### Rate Limiting

Rate limiting uses a **token bucket** algorithm implemented as an atomic Lua script running inside Redis. This ensures correctness even under high concurrency.

Four independent scopes are evaluated per request (most restrictive wins):

| Scope | Key Pattern | Description |
|-------|-------------|-------------|
| `global` | `limit:global` | Shared across all clients for a service |
| `ip` | `limit:ip:<client_ip>` | Per client IP address |
| `user` | `limit:user:<user_id>` | Per authenticated user (JWT only) |
| `key` | `limit:key:<api_key>` | Per API key (`X-API-Key` header) |

Each scope is configured with a **rate** (tokens replenished per second) and a **burst** (maximum bucket size). When a scope is exhausted, the gateway returns `429 Too Many Requests` with the `X-RateLimit-Scope` header indicating which bucket was hit.

The rate limiter is **fail-open** — if Redis is unreachable, requests are allowed through to prevent the gateway from becoming a single point of failure.

### IP Filtering

Each service can define an IP filter with one of two modes:

- **whitelist** — Only IPs in the list are allowed. All others receive `403 Forbidden`.
- **blacklist** — IPs in the list are blocked. All others are allowed.

The filter checks `X-Forwarded-For` first (for proxied deployments), falling back to the direct connection IP. Loopback addresses (`127.0.0.1`, `::1`) are always treated as listed.

### Logging

Every request produces a single structured JSON log line to stdout:

```json
{
  "time": "2026-02-15T12:00:00.000Z",
  "level": "INFO",
  "msg": "request_completed",
  "trace_id": "a1b2c3d4-...",
  "method": "GET",
  "path": "/my-service/api/v1/users",
  "status": 200,
  "ip": "127.0.0.1:54321",
  "latency": "2.345ms",
  "user_agent": "curl/8.0"
}
```

The `trace_id` is either taken from the incoming `X-Request-ID` header or generated as a new UUID. It is propagated to the backend service and returned to the client in the response.

### Reverse Proxy

Gate Keeper uses Go's `net/http/httputil.ReverseProxy` under the hood. One proxy instance is created per service at startup and reused for all requests to that service. The gateway rewrites the request path to the endpoint's configured path and forwards all original headers plus:

- `X-Request-ID` — Trace ID for distributed tracing.
- `X-User-ID` — Authenticated user ID (JWT endpoints only).

## Setup

### Prerequisites

- **Go 1.25+**
- **Redis** (for rate limiting)

### 1. Clone the repository

```bash
git clone https://github.com/your-username/gate-keeper.git
cd gate-keeper
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env` with your values:

```
REDIS_ADDRESS=localhost:6379
SERVER_PORT=8080
AUTH_TOKEN_SECRET=your-jwt-secret
CONFIG_PATH=configs/config.json
```

### 3. Configure your services

Edit `configs/config.json` to define the backend services and endpoints you want the gateway to manage. See `configs/config.example.json` for the format.

### 4. Build and run

```bash
# Using Make
make build
make run

# Or directly with Go
go build -o out/gate-keeper ./cmd/gate-keeper
./out/gate-keeper
```

### Using Docker Compose (recommended for local dev)

This starts both the gateway and a Redis instance:

```bash
docker compose up --build
```

### Makefile Targets

| Command | Description |
|---------|-------------|
| `make build` | Compile the binary to `out/gate-keeper` |
| `make run` | Build and run the gateway |
| `make test` | Run all tests |
| `make clean` | Remove build artifacts |
| `make lint` | Run `golangci-lint` |
