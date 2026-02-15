-- Updated Token Bucket Rate Limiter
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1

local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local last_tokens = tonumber(bucket[1]) or burst
local last_time = tonumber(bucket[2]) or now

local delta = math.max(0, now - last_time)
local replenished = delta * rate
local current_tokens = math.min(burst, last_tokens + replenished)

local allowed = 0
local remaining = current_tokens

if current_tokens >= requested then
    allowed = 1
    remaining = current_tokens - requested
    redis.call('HSET', key, 'tokens', remaining, 'last_time', now)
    redis.call('EXPIRE', key, 60) 
end

-- Return both the status and the remaining count
return {allowed, math.floor(remaining)}
