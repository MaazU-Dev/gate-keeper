package ratelimiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

// This directive tells Go to take the contents of the file
// and put it into this variable at build time.
//
//go:embed rate_limiter.lua
var luaScript string

type RateLimiter struct {
	rdb    *redis.Client
	script *redis.Script
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		script: redis.NewScript(luaScript),
	}
}

func (rl *RateLimiter) Check(ctx context.Context, key string, rate int, burst int) ([]int64, error) {
	now := time.Now().Unix()
	// two values returned by the script: {allowed, remaining}.
	values, err := rl.script.Run(ctx, rl.rdb, []string{key}, rate, burst, now).Int64Slice()
	if err != nil {
		// fail-open: if Redis is down, allow the traffic but log it
		return []int64{1, 0}, nil
	}
	if len(values) == 0 {
		// fail-open.
		return []int64{1, 0}, nil
	}
	return values, nil
}
