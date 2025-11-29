package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.client.Set(ctx, getJWTKey(jwtStr), true, jwtTTL).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	cmd := c.client.Get(ctx, getJWTKey(jwtStr))

	err := cmd.Err()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		log.WithError(err).WithField("jwtStr", jwtStr).Error("Error checking JWT in blacklist")
		return err // Propagate other Redis errors
	}

	log.WithField("jwtStr", jwtStr).Info("Found JWT in blacklist")
	return fmt.Errorf("JWT token is blacklisted")
}
