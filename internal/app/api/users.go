package api

import (
	"errors"
	"fmt"
	"net/http"
	"rip/internal/app/ds"
	"rip/internal/app/middlewares"
	"rip/internal/app/repositories"
	"rip/internal/app/services"
	"strings"

	"github.com/gin-gonic/gin"
)

type UsersApi struct {
	s *services.UsersService
}

func NewUsersApi(service *services.UsersService) *UsersApi {
	return &UsersApi{
		s: service,
	}
}

func (a *UsersApi) RegisterEndpoints(r *gin.RouterGroup, m *middlewares.UserMiddlewares) {
	r.POST("/", a.RegisterUser)
	r.GET("/", m.WithAuth, a.GetCurrentUser)
	r.PUT("/", m.WithAuth, a.UpdateCurrentUser)
	r.POST("/login/", a.Login)
	r.POST("/logout/", m.WithAuth, a.Logout)
}

// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя с указанными учетными данными
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param user body ds.CreateUser true "Данные нового пользователя"
// @Success 201 {object} ds.User "Созданный пользователь"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 409 {object} map[string]string "Логин уже занят"
// @Router /api/users/ [post]
func (a *UsersApi) RegisterUser(ctx *gin.Context) {
	user := ds.CreateUser{}
	if err := ctx.Bind(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	createdUser, err := a.s.RegisterUser(user)
	if err != nil {
		if errors.Is(err, repositories.ErrorLoginIsTaken) {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdUser)
}

// @Summary Получение информации о текущем пользователе
// @Description Возвращает информацию о текущем аутентифицированном пользователе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Success 200 {object} ds.User "Информация о пользователе"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Security JWT
// @Router /api/users/ [get]
func (a *UsersApi) GetCurrentUser(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	user, err := a.s.GetUser(userId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// @Summary Обновление информации о текущем пользователе
// @Description Обновляет информацию о текущем аутентифицированном пользователе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param user body ds.UpdateUser true "Данные для обновления"
// @Success 200 {object} ds.User "Обновленный пользователь"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 409 {object} map[string]string "Логин уже занят"
// @Security JWT
// @Router /api/users/ [put]
func (a *UsersApi) UpdateCurrentUser(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	user := ds.UpdateUser{}
	if err := ctx.Bind(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	updatedUser, err := a.s.UpdateUser(userId, user)
	if err != nil {
		if errors.Is(err, repositories.ErrorLoginIsTaken) {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updatedUser)
}

// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя и создает сессию
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param credentials body ds.CreateUser true "Учетные данные пользователя (логин и пароль)"
// @Success 200 {object} api.Login.Res "Сообщение об успешной аутентификации"
// @Failure 400 "Некорректные учетные данные"
// @Failure 403 "Доступ запрещен"
// @Failure 500 "Ошибка сервера"
// @Router /api/users/login/ [post]
func (a *UsersApi) Login(ctx *gin.Context) {
	type Req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	type Res struct {
		ExpiresIn   uint   `json:"expires_in"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	req := &Req{}
	if err := ctx.Bind(&req); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jwt, err := a.s.Authorize(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repositories.ErrorUserNotFound) || errors.Is(err, services.ErrorUserCredentialsIncorrect) {
			ctx.AbortWithError(http.StatusForbidden, fmt.Errorf("cant create str token"))
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
		}
		return
	}

	ctx.JSON(http.StatusOK, Res{
		ExpiresIn:   uint(a.s.Config.JWT.ExpiresIn.Seconds()),
		AccessToken: jwt,
		TokenType:   "Bearer",
	})
}

// @Summary Деавторизация пользователя
// @Description Завершает сессию текущего аутентифицированного пользователя, добавляя JWT в черный список
// @Tags Пользователи
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Сообщение об успешной деавторизации"
// @Success 200 {object} map[string]string "Сообщение об успешной деавторизации"
// @Security JWT
// @Router /api/users/logout/ [post]
func (a *UsersApi) Logout(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, middlewares.JwtPrefix) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtStr = jwtStr[len(middlewares.JwtPrefix):]

	a.s.Deauthorize(ctx, jwtStr)

	ctx.Status(http.StatusOK)
}
