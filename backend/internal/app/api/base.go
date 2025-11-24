package api

import (
	"fmt"
	"rip/internal/app/config"
	"rip/internal/app/ds"
	"rip/internal/app/middlewares"
	"rip/internal/app/redis"
	"rip/internal/app/repositories"
	"rip/internal/app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TurbinesAppApi struct {
	config *config.Config
	db     *gorm.DB
	redis  *redis.Client
}

func NewTurbinesAppApi(config *config.Config, db *gorm.DB, redis *redis.Client) *TurbinesAppApi {
	return &TurbinesAppApi{
		config: config,
		db:     db,
		redis:  redis,
	}
}

func (a *TurbinesAppApi) RegisterEndpoints(router *gin.RouterGroup) {
	userMiddlewares := middlewares.NewUserMiddlewares(a.config.JWT.Token, a.redis)

	generationRequestsService := services.NewGenerationRequestsService(a.db)
	generationRequestApi := NewGenerationRequestsApi(generationRequestsService)
	generationRequestApi.RegisterEndpoints(router.Group("/generation-requests"), &userMiddlewares)

	turbinesImagesS3, _ := repositories.NewS3Repository(
		a.config.S3.Host,
		a.config.S3.Port,
		"turbines",
	)
	turbinesService := services.NewTurbinesService(a.db, turbinesImagesS3)
	turbinesApi := NewTurbinesApi(turbinesService)
	turbinesApi.RegisterEndpoints(router.Group("/turbines"), &userMiddlewares)

	usersService := services.NewUsersService(a.db, a.config, a.redis)
	usersApi := NewUsersApi(usersService)
	usersApi.RegisterEndpoints(router.Group("/users"), &userMiddlewares)
}

func GetUserID(ctx *gin.Context) (uint, error) {
	userId, exists := ctx.Get(ds.UserIDKey)
	if !exists {
		return 0, fmt.Errorf("user claims expected but not found")
	}

	return userId.(uint), nil
}

func GetIsModerator(ctx *gin.Context) (bool, error) {
	isModerator, exists := ctx.Get(ds.UserIsModeratorKey)
	if !exists {
		return false, fmt.Errorf("user claims expected but not found")
	}

	return isModerator.(bool), nil
}
