package repositories

import (
	"errors"
	"rip/internal/app/ds"
	"time"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GenerationRequestRepository struct {
	GenerationRequestDB *gorm.DB
}

type GenerationRequestsFilter struct {
	Status        *string    `form:"status" binding:"omitnil,oneof=sent completed rejected"`
	FormedAtBegin *time.Time `form:"formed_at_begin" binding:"omitnil" time_format:"2006-01-02T15:04:05Z07:00"`
	FormedAtEnd   *time.Time `form:"formed_at_end" binding:"omitnil" time_format:"2006-01-02T15:04:05Z07:00"`
}

var (
	ErrorGenerationRequestNotFound              = errors.New("generation request was not found")
	ErrorGenerationRequestTurbineNotFound       = errors.New("generation request turbine was not found")
	ErrorGenerationRequestCannotBeClosed        = errors.New("generation request cannot be closed")
	ErrorGenerationRequestTurbineAlreadyInDraft = errors.New("generation request turbine is already in draft")
	ErrorGenerationRequestNotFilled             = errors.New("generation request is not filled properly")
	ErrorGenerationRequestTurbinesNotFilled     = errors.New("generation request turbines are not filled properly")
)

func NewGenerationRequestRepository(db *gorm.DB) *GenerationRequestRepository {
	return &GenerationRequestRepository{
		GenerationRequestDB: db,
	}
}

func (r *GenerationRequestRepository) GetGenerationRequests(userId *uint, filter GenerationRequestsFilter) ([]ds.GenerationRequest, error) {
	generationRequests := []ds.GenerationRequest{}

	query := r.GenerationRequestDB.Model(&ds.GenerationRequest{})

	if userId != nil {
		query = query.Where(ds.GenerationRequest{CreatedByID: *userId})
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(filter); err != nil {
		log.WithError(err).WithField("Filter", filter).Error("Failed to validate GenerationRequest filter")
		return []ds.GenerationRequest{}, err
	}

	if filter.Status != nil {
		query = query.Where(ds.GenerationRequest{Status: *filter.Status})
	} else {
		query = query.Where("status != ?", "deleted").Where("status != ?", "draft")
	}
	if filter.FormedAtBegin != nil {
		query = query.Where("formed_at > ?", *filter.FormedAtBegin)
	}
	if filter.FormedAtEnd != nil {
		query = query.Where("formed_at < ?", *filter.FormedAtEnd)
	}

	if err := query.Preload("CreatedBy").Preload("ClosedBy").Preload("TurbineGenerationRequests").Find(&generationRequests).Error; err != nil {
		log.WithError(err).Error("Failed to get generation requests from DB")
		return []ds.GenerationRequest{}, err
	}

	return generationRequests, nil
}

func (r *GenerationRequestRepository) GetGenerationRequest(generationRequestId uint) (ds.GenerationRequest, error) {
	generationRequest := ds.GenerationRequest{}

	if err := r.GenerationRequestDB.Where(&ds.GenerationRequest{
		ID: generationRequestId,
	}).Where("status != ?", "deleted").Preload("CreatedBy").Preload("ClosedBy").Preload("TurbineGenerationRequests.Turbine").First(&generationRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.GenerationRequest{}, ErrorGenerationRequestNotFound
		}

		log.WithError(err).WithFields(log.Fields{
			"Generation Request ID": generationRequest,
		}).Errorf("Failed to get generation request from DB")
		return ds.GenerationRequest{}, err
	}

	counter := uint(0)
	generationRequest.TurbineGenerationRequestsCount = &counter

	if generationRequest.TurbineGenerationRequests != nil && len(*generationRequest.TurbineGenerationRequests) > 0 {
		generationSum := uint64(0)
		for _, turbine := range *generationRequest.TurbineGenerationRequests {
			counter += 1

			if turbine.CalculatedGeneration != nil {
				generationSum += *turbine.CalculatedGeneration
			} else {
				log.Debugf("CalculatedGeneration is nil for Turbine ID %d in GenerationRequest ID %d, skipping.", turbine.Turbine.ID, generationRequest.ID)
				continue
			}
		}

		generationRequest.TurbineGenerationRequestsCount = &counter
		generationRequest.CalculatedGenerationSum = &generationSum
	}
	return generationRequest, nil
}

func (r *GenerationRequestRepository) GetGenerationRequestTurbines(generationRequestId uint) ([]ds.TurbineGenerationRequest, error) {
	generationRequestTurbines := []ds.TurbineGenerationRequest{}

	if err := r.GenerationRequestDB.Model(&ds.TurbineGenerationRequest{}).Where(&ds.TurbineGenerationRequest{
		GenerationRequestID: generationRequestId,
	}).Preload("Turbine", "is_active = ?", true).Find(&generationRequestTurbines).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Generation Request ID": generationRequestId,
		}).Errorf("Failed to get generation request from DB")
		return []ds.TurbineGenerationRequest{}, err
	}
	return generationRequestTurbines, nil
}

func (r *GenerationRequestRepository) GetGenerationRequestTurbine(generationRequestId, turbineId uint) (ds.TurbineGenerationRequest, error) {
	generationRequestTurbine := ds.TurbineGenerationRequest{}

	if err := r.GenerationRequestDB.Model(&ds.TurbineGenerationRequest{}).Where(&ds.TurbineGenerationRequest{
		GenerationRequestID: generationRequestId,
		TurbineID:           turbineId,
	}).First(&generationRequestTurbine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.TurbineGenerationRequest{}, ErrorGenerationRequestTurbineNotFound
		}
		log.WithError(err).WithFields(log.Fields{
			"Generation Request ID": generationRequestId,
		}).Errorf("Failed to get generation request from DB")
		return ds.TurbineGenerationRequest{}, err
	}
	return generationRequestTurbine, nil
}

func (r *GenerationRequestRepository) UpdateGenerationRequestTurbine(generationRequestId, turbineId uint, turbineGenerationRequestUpdates ds.UpdateTurbineGenerationRequest) (ds.TurbineGenerationRequest, error) {
	generationRequestTurbine, err := r.GetGenerationRequestTurbine(generationRequestId, turbineId)
	if err != nil {
		return ds.TurbineGenerationRequest{}, err
	}

	updates := map[string]any{}
	if turbineGenerationRequestUpdates.Alpha != nil {
		updates["alpha"] = *turbineGenerationRequestUpdates.Alpha
	}
	if turbineGenerationRequestUpdates.AvgVelocity != nil {
		updates["avg_velocity"] = *turbineGenerationRequestUpdates.AvgVelocity
	}
	if turbineGenerationRequestUpdates.CalculatedGeneration != nil {
		updates["calculated_generation"] = *turbineGenerationRequestUpdates.CalculatedGeneration
	}

	if len(updates) == 0 {
		return generationRequestTurbine, ErrorNothingToUpdate
	}

	updatedGenerationRequestTurbine := ds.TurbineGenerationRequest{}
	err = r.GenerationRequestDB.Model(&ds.TurbineGenerationRequest{}).Where(&ds.TurbineGenerationRequest{
		GenerationRequestID: generationRequestId,
		TurbineID:           turbineId,
	}).Clauses(clause.Returning{}).Updates(updates).Scan(&updatedGenerationRequestTurbine).Error
	if err != nil {
		return ds.TurbineGenerationRequest{}, err
	}
	return updatedGenerationRequestTurbine, nil
}

func (r *GenerationRequestRepository) UpdateDraftGenerationRequest(userId uint, generationRequestUpdates ds.UpdateGenerationRequest) (ds.GenerationRequest, error) {
	generationRequest, err := r.GetDraftGenerationRequest(userId)
	if err != nil {
		return ds.GenerationRequest{}, err
	}

	updates := map[string]any{}
	if generationRequestUpdates.PeriodDays != nil {
		updates["period_days"] = *generationRequestUpdates.PeriodDays
	}

	if len(updates) == 0 {
		return generationRequest, ErrorNothingToUpdate
	}

	updatedGenerationRequest := ds.GenerationRequest{}
	err = r.GenerationRequestDB.
		Model(&ds.GenerationRequest{}).
		Where("id = ?", generationRequest.ID).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&updatedGenerationRequest).Error
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Generation Request ID": generationRequest.ID,
			"Updates":               updates,
		}).Error("Failed to update generation request in DB")
		return ds.GenerationRequest{}, err
	}

	return updatedGenerationRequest, nil
}

func (r *GenerationRequestRepository) CompleteGenerationRequest(generationRequestId uint, moderatorId uint, calculationFn func(avgVelocity float32, height uint16, alpha float32, power uint32, days uint) uint64) (ds.GenerationRequest, error) {
	generationRequest, err := r.GetGenerationRequest(generationRequestId)
	if err != nil {
		return ds.GenerationRequest{}, err
	}
	if generationRequest.Status != "sent" {
		return ds.GenerationRequest{}, ErrorGenerationRequestCannotBeClosed
	}
	if generationRequest.PeriodDays == nil {
		return ds.GenerationRequest{}, ErrorGenerationRequestCannotBeClosed
	}

	transaction := r.GenerationRequestDB.Begin()
	transactionalService := NewGenerationRequestRepository(transaction)

	now := time.Now()
	if err := transaction.Where(&ds.GenerationRequest{
		ID: generationRequestId,
	}).Updates(&ds.GenerationRequest{
		Status:     "completed",
		ClosedAt:   &now,
		ClosedByID: &moderatorId,
	}).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{"Generation Request ID": generationRequestId}).Error("Failed to update generation request in DB")
		transaction.Rollback()
		return ds.GenerationRequest{}, err
	}

	turbines, err := transactionalService.GetGenerationRequestTurbines(generationRequestId)
	if err != nil {
		transaction.Rollback()
		return ds.GenerationRequest{}, err
	}

	for _, turbine := range turbines {
		calculatedGeneration := calculationFn(
			*turbine.AvgVelocity,
			turbine.Turbine.Height,
			*turbine.Alpha,
			turbine.Turbine.Power,
			*generationRequest.PeriodDays,
		)

		if err := transaction.Where(&ds.TurbineGenerationRequest{
			TurbineID:           turbine.TurbineID,
			GenerationRequestID: turbine.GenerationRequestID,
		}).Updates(&ds.TurbineGenerationRequest{
			CalculatedGeneration: &calculatedGeneration,
		}).Error; err != nil {
			log.WithError(err).WithFields(log.Fields{"Generation Request ID": generationRequestId}).Error("Failed to fill calculated generation in DB")
			transaction.Rollback()
			return ds.GenerationRequest{}, err
		}
	}

	transaction.Commit()
	return r.GetGenerationRequest(generationRequestId)
}

func (r *GenerationRequestRepository) RejectGenerationRequest(generationRequestId uint, moderatorId uint) (ds.GenerationRequest, error) {
	generationRequest, err := r.GetGenerationRequest(generationRequestId)
	if err != nil {
		return ds.GenerationRequest{}, nil
	}
	if generationRequest.Status != "sent" {
		return ds.GenerationRequest{}, ErrorGenerationRequestCannotBeClosed
	}

	now := time.Now()
	if err := r.GenerationRequestDB.Where(&ds.GenerationRequest{
		ID: generationRequestId,
	}).Updates(&ds.GenerationRequest{
		Status:     "rejected",
		ClosedAt:   &now,
		ClosedByID: &moderatorId,
	}).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{"Generation Request ID": generationRequestId}).Error("Failed to delete draft generation request from DB")
		return ds.GenerationRequest{}, err
	}
	return r.GetGenerationRequest(generationRequestId)
}

func (r *GenerationRequestRepository) GetDraftGenerationRequest(userId uint) (ds.GenerationRequest, error) {
	generationRequest := ds.GenerationRequest{}

	if err := r.GenerationRequestDB.Model(&ds.GenerationRequest{}).Where(&ds.GenerationRequest{
		CreatedByID: userId,
		Status:      "draft",
	}).Preload("TurbineGenerationRequests.Turbine").First(&generationRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.GenerationRequest{}, ErrorGenerationRequestNotFound
		}
		log.WithError(err).WithFields(log.Fields{
			"User ID": userId,
		}).Errorf("Failed to get draft generation request from DB")
		return ds.GenerationRequest{}, err
	}
	return generationRequest, nil
}

func (r *GenerationRequestRepository) GetOrCreateDraftGenerationRequest(userId uint) (ds.GenerationRequest, error) {
	generationRequest := ds.GenerationRequest{}

	if err := r.GenerationRequestDB.Where(&ds.GenerationRequest{
		CreatedByID: userId,
		Status:      "draft",
	}).Preload("TurbineGenerationRequests.Turbine").First(&generationRequest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newDraft := ds.GenerationRequest{
				CreatedByID: userId,
				Status:      "draft",
				CreatedAt:   time.Now(),
			}
			if err := r.GenerationRequestDB.Create(&newDraft).Error; err != nil {
				log.WithError(err).WithFields(log.Fields{
					"User ID": userId,
				}).Error("Failed to create draft generation request in DB")
				return ds.GenerationRequest{}, err
			}
			return newDraft, nil
		}
		log.WithError(err).WithFields(log.Fields{
			"User ID": userId,
		}).Error("Failed to get or create draft generation request from DB")
		return ds.GenerationRequest{}, err
	}

	return generationRequest, nil
}

func (r *GenerationRequestRepository) GetGenerationRequestsTurbinesCount(generationRequestId uint) (uint, error) {
	generationRequest := &ds.GenerationRequest{}
	err := r.GenerationRequestDB.Where(&ds.GenerationRequest{
		ID: generationRequestId,
	}).First(generationRequest).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	count := r.GenerationRequestDB.Model(generationRequest).Association("TurbineGenerationRequests").Count()
	return uint(count), nil
}

func (r *GenerationRequestRepository) AddTurbineToDraftGenerationRequest(generationRequestId uint, turbineId uint) error {
	err := r.GenerationRequestDB.Where(&ds.TurbineGenerationRequest{
		TurbineID:           turbineId,
		GenerationRequestID: generationRequestId,
	}).First(&ds.TurbineGenerationRequest{}).Error

	if err == nil {
		return ErrorGenerationRequestTurbineAlreadyInDraft
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.GenerationRequestDB.Create(&ds.TurbineGenerationRequest{
			TurbineID:           turbineId,
			GenerationRequestID: generationRequestId,
		}).Error
	}

	return err
}

func (r *GenerationRequestRepository) UpdateTurbineInDraft(generationRequestId uint, turbineId uint, turbineUpdates ds.UpdateTurbineGenerationRequest) error {
	updates := map[string]any{}
	if turbineUpdates.Alpha != nil {
		updates["alpha"] = *turbineUpdates.Alpha
	}
	if turbineUpdates.AvgVelocity != nil {
		updates["avg_velocity"] = *turbineUpdates.AvgVelocity
	}

	err := r.GenerationRequestDB.
		Model(&ds.TurbineGenerationRequest{}).
		Where(ds.TurbineGenerationRequest{
			TurbineID:           turbineId,
			GenerationRequestID: generationRequestId,
		}).
		Clauses(clause.Returning{}).
		Updates(updates).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorGenerationRequestTurbineNotFound
	}
	return err
}

func (r *GenerationRequestRepository) RemoveTurbineFromDraftGenerationRequest(generationRequestId uint, turbineId uint) error {
	if err := r.GenerationRequestDB.Where(&ds.TurbineGenerationRequest{
		GenerationRequestID: generationRequestId,
		TurbineID:           turbineId,
	}).Delete(&ds.TurbineGenerationRequest{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorGenerationRequestTurbineNotFound
		}
		log.WithError(err).WithFields(log.Fields{
			"Generation Request ID": generationRequestId,
			"Turbine ID":            turbineId,
		}).Error("Failed to remove turbine from draft generation request in DB")
		return err
	}
	return nil
}

func (r *GenerationRequestRepository) SubmitDraftGenerationRequest(userId uint) error {
	currentDraft, err := r.GetDraftGenerationRequest(userId)
	if err != nil {
		return err
	}

	draftTurbines, err := r.GetGenerationRequestTurbines(currentDraft.ID)
	if err != nil {
		return err
	}

	if currentDraft.PeriodDays == nil {
		return ErrorGenerationRequestNotFilled
	}

	for _, turbine := range draftTurbines {
		if turbine.Alpha == nil || turbine.AvgVelocity == nil {
			return ErrorGenerationRequestTurbinesNotFilled
		}
	}

	now := time.Now()
	if err := r.GenerationRequestDB.Debug().Model(&ds.GenerationRequest{}).Where(&ds.GenerationRequest{
		ID: currentDraft.ID,
	}).Updates(&ds.GenerationRequest{
		Status:   "sent",
		FormedAt: &now,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorGenerationRequestNotFound
		}
		return err
	}
	return nil
}

func (r *GenerationRequestRepository) DeleteGenerationRequest(generationRequestId uint) error {
	now := time.Now()
	if err := r.GenerationRequestDB.Where(&ds.GenerationRequest{
		ID: generationRequestId,
	}).Updates(&ds.GenerationRequest{
		Status:   "deleted",
		FormedAt: &now,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorGenerationRequestNotFound
		}
		log.WithError(err).WithFields(log.Fields{"Generation Request ID": generationRequestId}).Error("Failed to delete draft generation request from DB")
		return err
	}
	return nil
}
