package api

import (
	"errors"
	"net/http"
	"rip/internal/app/ds"
	"rip/internal/app/middlewares"
	"rip/internal/app/repositories"
	"rip/internal/app/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TurbinesApi struct {
	s *services.TurbinesService
}

func NewTurbinesApi(s *services.TurbinesService) *TurbinesApi {
	return &TurbinesApi{
		s: s,
	}
}

func (a *TurbinesApi) RegisterEndpoints(r *gin.RouterGroup, m *middlewares.UserMiddlewares) {
	r.GET("/", a.GetTurbines)

	r.POST("/", m.WithModeratorAccess, a.CreateTurbine)
	r.GET("/:turbineId/", a.GetTurbine)
	r.PUT("/:turbineId/", m.WithModeratorAccess, a.UpdateTurbine)
	r.POST("/:turbineId/upload-image/", m.WithModeratorAccess, a.UploadTurbineImage)
	r.DELETE("/:turbineId/", m.WithModeratorAccess, a.DeleteTurbine)
}

// @Summary Список турбин с фильтрацией
// @Description Возвращает список активных турбин с возможностью фильтрации по названию
// @Tags Ветрогенераторы
// @Accept json
// @Produce json
// @Param turbineTitle query string false "Фильтр по названию турбины"
// @Success 200 {array} ds.Turbine "Список турбин"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/turbines/ [get]
func (a *TurbinesApi) GetTurbines(ctx *gin.Context) {
	titleFilter := func() *string {
		titleFilterValue := ctx.Query("turbineTitle")
		if titleFilterValue != "" {
			return &titleFilterValue
		}
		return nil
	}()

	turbines, err := a.s.GetActiveTurbines(titleFilter)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, turbines)
}

// @Summary Получение информации о турбине
// @Description Возвращает информацию о конкретной турбине по ID
// @Tags Ветрогенераторы
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Success 200 {object} ds.Turbine "Информация о турбине"
// @Failure 404 {object} map[string]string "Турбина не найдена"
// @Failure 422 {object} map[string]string "Некорректный ID турбины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/turbines/{turbineId}/ [get]
func (a *TurbinesApi) GetTurbine(ctx *gin.Context) {
	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}

	turbine, err := a.s.GetActiveTurbine(uint(turbineId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, turbine)
}

// @Summary Создание новой турбины
// @Description Создает новую турбину без изображения
// @Tags Ветрогенераторы
// @Accept json
// @Produce json
// @Param turbine body ds.CreateTurbine true "Данные новой турбины"
// @Success 201 {object} ds.Turbine "Созданная турбина"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/turbines/ [post]
func (a *TurbinesApi) CreateTurbine(ctx *gin.Context) {
	turbine := ds.CreateTurbine{}

	if err := ctx.Bind(&turbine); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	createdTurbine, err := a.s.CreateTurbine(turbine)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdTurbine)
}

// @Summary Обновление информации о турбине
// @Description Обновляет информацию о турбине по ID
// @Tags Ветрогенераторы
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Param turbine body ds.UpdateTurbine true "Данные для обновления"
// @Success 200 {object} ds.Turbine "Обновленная турбина"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 404 {object} map[string]string "Турбина не найдена"
// @Failure 422 {object} map[string]string "Некорректный ID турбины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/turbines/{turbineId}/ [put]
func (a *TurbinesApi) UpdateTurbine(ctx *gin.Context) {
	turbine := ds.UpdateTurbine{}
	if err := ctx.Bind(&turbine); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}

	updatedTurbine, err := a.s.UpdateTurbine(uint(turbineId), turbine)
	if err != nil {
		if errors.Is(err, repositories.ErrorTurbineIsDeleted) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTurbine)
}

// @Summary Добавление изображения к турбине
// @Description Добавляет или обновляет изображение для турбины
// @Tags Ветрогенераторы
// @Accept multipart/form-data
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Param image formData file true "Файл изображения"
// @Success 200 {object} ds.Turbine "Обновленная турбина"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 404 {object} map[string]string "Турбина не найдена"
// @Failure 422 {object} map[string]string "Некорректный ID турбины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/turbines/{turbineId}/upload-image/ [post]
func (a *TurbinesApi) UploadTurbineImage(ctx *gin.Context) {
	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}

	_, err = a.s.GetActiveTurbine(uint(turbineId))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	if ctx.Request.ContentLength <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	updatedTurbine, err := a.s.UpdateTurbineImage(
		uint(turbineId),
		ctx.Request.Body,
		ctx.Request.ContentLength,
		ctx.ContentType(),
	)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTurbine)
}

// @Summary Удаление турбины
// @Description Удаляет турбину по ID
// @Tags Ветрогенераторы
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Success 204 "Турбина успешно удалена"
// @Failure 404 {object} map[string]string "Турбина не найдена"
// @Failure 422 {object} map[string]string "Некорректный ID турбины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/turbines/{turbineId}/ [delete]
func (a *TurbinesApi) DeleteTurbine(ctx *gin.Context) {
	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}

	if err := a.s.DeleteTurbine(uint(turbineId)); err != nil {
		if errors.Is(err, repositories.ErrorTurbineIsDeleted) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
