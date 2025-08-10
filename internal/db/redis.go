package db

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedis(ctx context.Context, addr string) (*Redis, error) {
	if addr == "" {
		addr = "localhost:6379"
	}
	r := redis.NewClient(&redis.Options{Addr: addr})
	if err := r.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: r, ctx: ctx}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) SetOTP(phone, otp string, ttl time.Duration) error {
	return r.client.Set(r.ctx, "otp:"+phone, otp, ttl).Err()
}

func (r *Redis) GetOTP(phone string) (string, error) {
	return r.client.Get(r.ctx, "otp:"+phone).Result()
}

func (r *Redis) DeleteOTP(phone string) error {
	return r.client.Del(r.ctx, "otp:"+phone).Err()
}

func (r *Redis) IncrementRate(phone string, window time.Duration) (int64, error) {
	key := "rl:" + phone
	n, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = r.client.Expire(r.ctx, key, window).Err()
	}
	return n, nil
}

func (r *Redis) GetTTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

func (r *Redis) GetRate(phone string) (int64, error) {
	key := "rl:" + phone
	n, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(n, 10, 64)
}
