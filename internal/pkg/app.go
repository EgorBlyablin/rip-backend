package pkg

import (
	"fmt"
	"time"

	"rip/internal/app/api"
	"rip/internal/app/config"
	"rip/internal/app/redis"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "rip/docs"
)

type TurbinesApplication struct{}

func (a *TurbinesApplication) Run(config *config.Config, db *gorm.DB, redis *redis.Client) {
	logrus.Info("Server start up")

	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"https://egorblyablin.github.io",
			"https://10.185.38.237:3000",
			"https://localhost:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router := engine.Group("/api")

	turbinesApi := api.NewTurbinesAppApi(config, db, redis)
	turbinesApi.RegisterEndpoints(router)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	serverAddress := fmt.Sprintf("%s:%d", config.Service.Host, config.Service.Port)
	if err := engine.RunTLS(serverAddress, "/app/cert.crt", "/app/cert.key"); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Server down")
}
