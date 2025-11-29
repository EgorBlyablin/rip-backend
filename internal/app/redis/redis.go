package redis

import (
	"context"
	"fmt"
	"rip/internal/app/config"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "turbines_api."

type Client struct {
	cfg    config.Config
	client *redis.Client
}

func New(ctx context.Context, cfg config.Config) (*Client, error) {
	client := &Client{}

	client.cfg = cfg

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfg.Redis.Password,
		Username:    cfg.Redis.User,
		Addr:        cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		DB:          0,
		DialTimeout: cfg.Redis.DialTimeout,
		ReadTimeout: cfg.Redis.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
