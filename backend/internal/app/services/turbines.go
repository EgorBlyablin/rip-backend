package services

import (
	"io"
	"rip/internal/app/ds"
	"rip/internal/app/repositories"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TurbinesService struct {
	r      *repositories.TurbinesRepository
	images *repositories.S3Repository
}

func NewTurbinesService(db *gorm.DB, s3 *repositories.S3Repository) *TurbinesService {
	return &TurbinesService{
		r:      repositories.NewTurbinesRepository(db),
		images: s3,
	}
}

func (s *TurbinesService) GetActiveTurbines(titleFilter *string) ([]ds.Turbine, error) {
	return s.r.GetActiveTurbines(repositories.TurbinesFilter{
		Title: titleFilter,
	})
}

func (s *TurbinesService) GetActiveTurbine(turbineId uint) (ds.Turbine, error) {
	return s.r.GetActiveTurbine(turbineId)
}

func (s *TurbinesService) CreateTurbine(turbine ds.CreateTurbine) (ds.Turbine, error) {
	return s.r.CreateTurbine(turbine)
}

func (s *TurbinesService) UpdateTurbine(turbineId uint, turbine ds.UpdateTurbine) (ds.Turbine, error) {
	_, err := s.GetActiveTurbine(turbineId)
	if err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to get turbine during update process")
		return ds.Turbine{}, err
	}

	return s.r.UpdateTurbine(turbineId, turbine)
}

func (s *TurbinesService) UpdateTurbineImage(turbineId uint, image io.Reader, imageSize int64, imageContentType string) (ds.Turbine, error) {
	turbine, err := s.GetActiveTurbine(turbineId)
	if err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to get turbine during image upload process")
		return ds.Turbine{}, err
	}

	turbineImageKey, err := s.images.UploadFile(turbine.ImageSlug(), image, imageSize, imageContentType)
	if err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to upload turbine image")
		return ds.Turbine{}, err
	}

	updatedTurbine, err := s.r.UpdateTurbine(turbineId, ds.UpdateTurbine{
		Image: &turbineImageKey,
	})
	return updatedTurbine, err
}

func (s *TurbinesService) DeleteTurbine(turbineId uint) error {
	turbine, err := s.GetActiveTurbine(turbineId)
	if err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to get turbine during deletion process")
		return err
	}

	if turbine.Image != nil {
		err = s.images.DeleteFile(turbine.ImageSlug())
		if err != nil {
			log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to delete turbine image")
			return err
		}
	}

	err = s.r.DeleteTurbine(turbineId)
	if err != nil {
		log.WithError(err).WithField("Turbine ID", turbineId).Error("Failed to delete turbine")
		return err
	}

	return nil
}
