package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client      *redis.Client
	maxRequests int
	window      time.Duration
}

func NewRedisRateLimiter(redisAddr string, maxRequests int, window time.Duration) *RedisRateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisRateLimiter{
		client:      rdb,
		maxRequests: maxRequests,
		window:      window,
	}
}

// AllowRequest verifica si la IP actual tiene permitido realizar la petición
func (rl *RedisRateLimiter) AllowRequest(ctx context.Context, ip string) (bool, int, error) {
	key := fmt.Sprintf("rate_limit:%s", ip)

	// Incrementar el contador en Redis
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	// Si es la primera petición en esta ventana, establecemos el tiempo de expiración
	if count == 1 {
		rl.client.Expire(ctx, key, rl.window)
	}

	// Si el contador supera el máximo permitido
	if int(count) > rl.maxRequests {
		return false, int(count), nil
	}

	return true, int(count), nil
}
