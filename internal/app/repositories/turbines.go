package repositories

import (
	"errors"
	"rip/internal/app/ds"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func boolPtr(b bool) *bool { return &b }

type TurbinesRepository struct {
	turbinesDB *gorm.DB
}

type TurbinesFilter struct {
	Title *string
}

var (
	ErrorTurbineIsDeleted = errors.New("specified turbine is deleted")
	ErrorTurbineNotFound  = errors.New("turbine was not found")
)

func NewTurbinesRepository(db *gorm.DB) *TurbinesRepository {
	return &TurbinesRepository{
		turbinesDB: db,
	}
}

func (r *TurbinesRepository) GetActiveTurbines(filter TurbinesFilter) ([]ds.Turbine, error) {
	return r.getTurbines(filter, true)
}

func (r *TurbinesRepository) GetTurbines(filter TurbinesFilter) ([]ds.Turbine, error) {
	return r.getTurbines(filter, false)
}

func (r *TurbinesRepository) getTurbines(filter TurbinesFilter, excludeDeleted bool) ([]ds.Turbine, error) {
	turbines := []ds.Turbine{}

	query := r.turbinesDB

	if filter.Title != nil {
		query = query.Where("title ILIKE ?", "%"+*filter.Title+"%")
	}
	if excludeDeleted {
		query = query.Where(ds.Turbine{IsActive: true})
	}

	if err := query.Find(&turbines).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Title filter": filter.Title,
		}).Error("Failed to get turbines from DB")
		return []ds.Turbine{}, err
	}

	return turbines, nil
}

func (r *TurbinesRepository) GetActiveTurbine(turbineId uint) (ds.Turbine, error) {
	turbine, err := r.GetTurbine(turbineId)
	if err != nil {
		return ds.Turbine{}, err
	}
	if !turbine.IsActive {
		return ds.Turbine{}, ErrorTurbineIsDeleted
	}

	return turbine, nil
}

func (r *TurbinesRepository) GetTurbine(turbineId uint) (ds.Turbine, error) {
	turbine := ds.Turbine{}

	if err := r.turbinesDB.Where(&ds.Turbine{ID: turbineId}).First(&turbine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Turbine{}, ErrorTurbineNotFound
		}
		log.WithError(err).WithFields(log.Fields{"Turbine ID": turbineId}).Error("Failed to get turbine from DB")
		return ds.Turbine{}, err
	}

	return turbine, nil
}

func (r *TurbinesRepository) CreateTurbine(turbine ds.CreateTurbine) (ds.Turbine, error) {
	createdTurbine := ds.Turbine{
		Title:       turbine.Title,
		Description: turbine.Description,
		IsActive:    turbine.IsActive,
		Power:       turbine.Power,
		Height:      turbine.Height,
	}

	if err := r.turbinesDB.Create(&createdTurbine).Error; err != nil {
		log.WithError(err).WithField("Turbine", turbine).Error("Failed to create turbine in DB")
		return ds.Turbine{}, err
	}

	return createdTurbine, nil
}

func (r *TurbinesRepository) UpdateTurbine(turbineId uint, turbine ds.UpdateTurbine) (ds.Turbine, error) {
	_, err := r.GetActiveTurbine(turbineId)
	if err != nil {
		return ds.Turbine{}, err
	}

	updatedTurbine := ds.Turbine{}

	updates := map[string]any{}
	if turbine.Title != nil {
		updates["title"] = *turbine.Title
	}
	if turbine.Description != nil {
		updates["description"] = *turbine.Description
	}
	if turbine.IsActive != nil {
		updates["is_active"] = *turbine.IsActive
	}
	if turbine.Image != nil {
		updates["image"] = *turbine.Image
	}
	if turbine.Power != nil {
		updates["power"] = *turbine.Power
	}
	if turbine.Height != nil {
		updates["height"] = *turbine.Height
	}

	query := r.turbinesDB.Model(&ds.Turbine{}).Where("id = ?", turbineId)
	if len(updates) > 0 {
		query = query.Updates(updates)
	} else {
		log.WithField("Turbine ID", turbineId).Info("No fields to update for turbine")
		return r.GetActiveTurbine(turbineId)
	}

	if err := query.Clauses(clause.Returning{}).Scan(&updatedTurbine).Error; err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Turbine ID": turbineId,
			"Turbine":    turbine,
		}).Error("Failed to update turbine in DB")
		return ds.Turbine{}, err
	}

	return updatedTurbine, nil
}

func (r *TurbinesRepository) DeleteTurbine(turbineId uint) error {
	_, err := r.GetActiveTurbine(turbineId)
	if err != nil {
		return err
	}

	if _, err := r.UpdateTurbine(turbineId, ds.UpdateTurbine{IsActive: boolPtr(false)}); err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to delete turbine from DB")
		return err
	}
	return nil
}
