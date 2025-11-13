package main

import (
	"context"
	"fmt"

	"rip/internal/app/config"
	"rip/internal/app/dsn"
	"rip/internal/app/redis"
	"rip/internal/app/repositories"
	"rip/internal/pkg"

	"github.com/sirupsen/logrus"
)

// @title Turbines API
// @version 0.2.0
// @description API для работы с сервисом расчета генерации электроэнергии ветрогенераторами
// @license.name MIT License
// @contact.name Егор Бляблин
// @host localhost:8000
// @schemes http
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
// @description Префикс "Bearer" с последующими пробелом и JWT.
func main() {
	config, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresDsn := dsn.FromEnv()
	fmt.Println(postgresDsn)

	db, err := repositories.NewTurbinesDB(postgresDsn)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	ctx := context.Background()
	redisClient, err := redis.New(ctx, config.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	app := &pkg.TurbinesApplication{}
	app.Run(config, db, redisClient)
}
