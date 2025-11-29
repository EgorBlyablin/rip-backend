package middlewares

import (
	"net/http"
	"rip/internal/app/ds"
	"rip/internal/app/redis"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
)

const JwtPrefix = "Bearer "

type UserMiddlewares struct {
	Token string
	redis *redis.Client
}

func NewUserMiddlewares(token string, redis *redis.Client) UserMiddlewares {
	return UserMiddlewares{
		Token: token,
		redis: redis,
	}
}

func (u *UserMiddlewares) WithOptionalAuth(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, JwtPrefix) {
		ctx.Next()
		return
	}

	jwtStr = jwtStr[len(JwtPrefix):]

	if u.redis.CheckJWTInBlacklist(ctx, jwtStr) != nil {
		log.WithField("jwtStr", jwtStr).Info("Found JWT in blacklist")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwt, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.Token), nil
	})
	if err != nil {
		log.WithError(err).Info("Failed to parse JWT")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwtData := jwt.Claims.(*ds.JWTClaims)
	ctx.Set(ds.UserIDKey, jwtData.UserID)
	ctx.Set(ds.UserIsModeratorKey, jwtData.IsModerator)
	ctx.Next()
}

func (u *UserMiddlewares) WithAuth(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, JwtPrefix) {
		log.Error("No Authorization header found")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtStr = jwtStr[len(JwtPrefix):]

	if u.redis.CheckJWTInBlacklist(ctx, jwtStr) != nil {
		log.WithField("jwtStr", jwtStr).Info("Found JWT in blacklist")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwt, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.Token), nil
	})
	if err != nil {
		log.WithError(err).Info("Failed to parse JWT")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwtData := jwt.Claims.(*ds.JWTClaims)
	ctx.Set(ds.UserIDKey, jwtData.UserID)
	ctx.Set(ds.UserIsModeratorKey, jwtData.IsModerator)
	ctx.Next()
}

func (u *UserMiddlewares) WithModeratorAccess(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, JwtPrefix) {
		log.Error("No Authorization header found")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtStr = jwtStr[len(JwtPrefix):]

	if u.redis.CheckJWTInBlacklist(ctx, jwtStr) != nil {
		log.WithField("jwtStr", jwtStr).Info("Found JWT in blacklist")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	jwt, _ := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.Token), nil
	})

	jwtData := jwt.Claims.(*ds.JWTClaims)
	if !jwtData.IsModerator {
		log.Error("User is not a moderator but tried to access moderator-only resource")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	ctx.Set(ds.UserIDKey, jwtData.UserID)
	ctx.Set(ds.UserIsModeratorKey, true)
	ctx.Next()
}
