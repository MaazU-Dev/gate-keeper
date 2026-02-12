local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1

local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local last_tokens = tonumber(bucket[1]) or burst
local last_time = tonumber(bucket[2]) or now

-- Calculate how many tokens were generated since last call
local delta = math.max(0, now - last_time)
local replenished = delta * rate
local current_tokens = math.min(burst, last_tokens + replenished)

if current_tokens >= requested then
    -- Allow request
    redis.call('HSET', key, 'tokens', current_tokens - requested, 'last_time', now)
    redis.call('EXPIRE', key, 60) -- Cleanup unused buckets
    return 1
else
    -- Deny request
    return 0
end