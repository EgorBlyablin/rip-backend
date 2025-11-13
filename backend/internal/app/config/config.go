package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Service struct {
		Host string
		Port int
	}

	S3 struct {
		Host string
		Port int
	}

	JWT struct {
		Token         string
		SigningMethod string
		ExpiresIn     time.Duration
	}

	Redis RedisConfig
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
)

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("failed to read config: %v", err)
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Errorf("failed to unmarshal config: %v", err)
		return nil, err
	}

	cfg.Redis.Host = os.Getenv(envRedisHost)

	redisPortStr := os.Getenv(envRedisPort)
	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil {
		return nil, fmt.Errorf("redis port must be int value: %w", err)
	}
	cfg.Redis.Port = redisPort

	cfg.Redis.Password = os.Getenv(envRedisPass)
	cfg.Redis.User = os.Getenv(envRedisUser)

	log.Info("config parsed")

	return cfg, nil
}
