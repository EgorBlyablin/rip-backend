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
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type GenerationRequestsApi struct {
	s *services.GenerationRequestsService
}

func NewGenerationRequestsApi(s *services.GenerationRequestsService) *GenerationRequestsApi {
	return &GenerationRequestsApi{
		s: s,
	}
}

func (a *GenerationRequestsApi) RegisterEndpoints(r *gin.RouterGroup, m *middlewares.UserMiddlewares) {
	r.GET("/", m.WithAuth, a.GetSentGenerationRequests)
	r.GET("/:generationRequestId", m.WithAuth, a.GetGenerationRequest)
	r.PUT("/:generationRequestId/close", m.WithModeratorAccess, a.CloseGenerationRequest)

	r.GET("/draft", m.WithOptionalAuth, a.GetDraftBriefInfo)
	r.POST("/draft/:turbineId", m.WithAuth, a.AddTurbineToDraft)
	r.PUT("/draft", m.WithAuth, a.UpdateDraftGenerationRequest)
	r.PUT("/draft/:turbineId", m.WithAuth, a.UpdateTurbineInDraft)
	r.DELETE("/draft/:turbineId", m.WithAuth, a.RemoveTurbineFromDraft)
	r.PUT("/draft/submit", m.WithAuth, a.SubmitDraftGenerationRequest)
	r.DELETE("/draft", m.WithAuth, a.DeleteDraftGenerationRequest)
}

// @Summary Список заявок с фильтрацией
// @Description Возвращает список заявок пользователя с фильтрацией по дате и статусу (кроме удаленных и черновиков)
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param filter path repositories.GenerationRequestsFilter false "Фильтр заявок"
// @Success 200 {array} ds.GenerationRequest "Список заявок"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/ [get]
func (a *GenerationRequestsApi) GetSentGenerationRequests(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	generationRequestsFilter := repositories.GenerationRequestsFilter{}
	if err := ctx.BindQuery(&generationRequestsFilter); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(generationRequestsFilter); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userIsModeratorStr, found := ctx.Get(ds.UserIsModeratorKey)
	userFilter := &userId

	if found {
		if userIsModeratorStr.(bool) {
			userFilter = nil
		}
	}

	generationRequests, err := a.s.GetGenerationRequests(userFilter, generationRequestsFilter)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for i := range generationRequests {
		if generationRequests[i].Status == "completed" {
			counter := uint(len(*generationRequests[i].TurbineGenerationRequests))
			generationRequests[i].TurbineGenerationRequestsCount = &counter
		}
		generationRequests[i].TurbineGenerationRequests = nil
	}

	ctx.JSON(http.StatusOK, generationRequests)
}

// @Summary Получение информации о заявке
// @Description Возвращает информацию о конкретной заявке с деталями и связанными турбинами
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param generationRequestId path int true "ID заявки"
// @Success 200 {object} ds.GenerationRequest "Информация о заявке"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/{generationRequestId}/ [get]
func (a *GenerationRequestsApi) GetGenerationRequest(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	generationRequestsId, err := strconv.Atoi(ctx.Param("generationRequestId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	generationRequest, err := a.s.GetGenerationRequest(uint(generationRequestsId))
	if err != nil {
		if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if generationRequest.CreatedByID != userId {
		isModerator, err := GetIsModerator(ctx)
		if err != nil || !isModerator {
			log.Error("User tried to access a generation request that does not belong to them")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
	}

	ctx.JSON(http.StatusOK, generationRequest)
}

// @Summary Завершение/отклонение заявки модератором
// @Description Модератор завершает или отклоняет заявку (только для модераторов)
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param generationRequestId path int true "ID заявки"
// @Param request body api.CloseGenerationRequest.req true "Статус: completed или rejected"
// @Success 200 {object} ds.GenerationRequest "Обновленная заявка"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 409 {object} map[string]string "Заявка не может быть закрыта"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/{generationRequestId}/close/ [put]
func (a *GenerationRequestsApi) CloseGenerationRequest(ctx *gin.Context) {
	moderatorId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	generationRequestsId, err := strconv.Atoi(ctx.Param("generationRequestId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	type req struct {
		Status string `json:"status" binding:"required,oneof=completed rejected"`
	}
	closeGenerationRequest := req{}
	if err := ctx.Bind(&closeGenerationRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	closedGenerationRequest, err := a.s.CloseGenerationRequest(uint(generationRequestsId), moderatorId, closeGenerationRequest.Status)
	if err != nil {
		switch err {
		case services.ErrorGenerationRequestIncorrectStatus:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		case repositories.ErrorGenerationRequestNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		case repositories.ErrorGenerationRequestCannotBeClosed:
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": err.Error()})
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, closedGenerationRequest)
}

// @Summary Получение информации о черновике заявки
// @Description Возвращает информацию о черновике заявки текущего пользователя
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Success 200 {object} ds.DraftGenerationRequestsBriefInfo "Информация о черновике"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/generation-requests/draft/ [get]
func (a *GenerationRequestsApi) GetDraftBriefInfo(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusOK, ds.DraftGenerationRequestsBriefInfo{
			GenerationRequestId: 0,
			TurbinesCount:       0,
		})
		return
	}

	generationRequestDraftBriefInfo, err := a.s.GetDraftBriefInfo(userId)
	if err != nil {
		if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
			ctx.AbortWithStatusJSON(http.StatusOK, ds.DraftGenerationRequestsBriefInfo{
				GenerationRequestId: 0,
				TurbinesCount:       0,
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, generationRequestDraftBriefInfo)
}

// @Summary Обновление черновика заявки
// @Description Обновляет поля черновика заявки текущего пользователя
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param request body ds.UpdateGenerationRequest true "Данные для обновления"
// @Success 200 {object} ds.GenerationRequest "Обновленный черновик"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Черновик не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/ [put]
func (a *GenerationRequestsApi) UpdateDraftGenerationRequest(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	generationRequestUpdates := ds.UpdateGenerationRequest{}
	if err := ctx.Bind(&generationRequestUpdates); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	updatedGenerationRequest, err := a.s.UpdateDraftGenerationRequest(userId, generationRequestUpdates)
	if err != nil {
		if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedGenerationRequest)
}

// @Summary Добавление турбины в черновик
// @Description Добавляет турбину в черновик заявки текущего пользователя
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Success 200 {object} map[string]string "Турбина добавлена"
// @Failure 304 {object} map[string]string "Турбина уже в черновике"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/{turbineId}/ [post]
func (a *GenerationRequestsApi) AddTurbineToDraft(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	turbinesId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = a.s.AddTurbineToDraft(userId, uint(turbinesId))
	if err != nil {
		if errors.Is(err, repositories.ErrorGenerationRequestTurbineAlreadyInDraft) {
			ctx.AbortWithStatusJSON(http.StatusNotModified, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// @Summary Обновление параметров турбины в черновике
// @Description Обновляет параметры (avg_velocity, alpha) турбины в черновике
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Param updates body ds.UpdateTurbineGenerationRequest true "Параметры для обновления"
// @Success 200 {object} map[string]string "Параметры обновлены"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Турбина не найдена в черновике"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/{turbineId}/ [put]
func (a *GenerationRequestsApi) UpdateTurbineInDraft(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	turbineUpdates := ds.UpdateTurbineGenerationRequest{}
	if err := ctx.Bind(&turbineUpdates); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = a.s.UpdateTurbineInDraft(userId, uint(turbineId), turbineUpdates)
	if errors.Is(err, repositories.ErrorGenerationRequestTurbineNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// @Summary Удаление турбины из черновика
// @Description Удаляет турбину из черновика заявки текущего пользователя
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Param turbineId path int true "ID турбины"
// @Success 200 {object} map[string]string "Турбина удалена"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Турбина или черновик не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/{turbineId}/ [delete]
func (a *GenerationRequestsApi) RemoveTurbineFromDraft(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	turbineId, err := strconv.Atoi(ctx.Param("turbineId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = a.s.RemoveTurbineFromDraft(userId, uint(turbineId))
	if errors.Is(err, repositories.ErrorGenerationRequestTurbineNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// @Summary Отправка черновика заявки
// @Description Формирует черновик заявки и отправляет на рассмотрение
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Черновик отправлен"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Черновик не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/submit/ [put]
func (a *GenerationRequestsApi) SubmitDraftGenerationRequest(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	err = a.s.SubmitDraftGenerationRequest(userId)
	if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// @Summary Удаление черновика заявки
// @Description Удаляет черновик заявки текущего пользователя
// @Tags Заявки расчета выработки
// @Accept json
// @Produce json
// @Success 204 "Черновик удален"
// @Failure 404 {object} map[string]string "Черновик не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security JWT
// @Router /api/generation-requests/draft/ [delete]
func (a *GenerationRequestsApi) DeleteDraftGenerationRequest(ctx *gin.Context) {
	userId, err := GetUserID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	err = a.s.DeleteDraftGenerationRequest(userId)
	if err != nil {
		if errors.Is(err, repositories.ErrorGenerationRequestNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
