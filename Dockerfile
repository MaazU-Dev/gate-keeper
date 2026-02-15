# ── Build stage ────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gate-keeper ./cmd/gate-keeper

# ── Runtime stage ─────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /gate-keeper .
COPY configs/ ./configs/

EXPOSE 8080

CMD ["./gate-keeper"]
