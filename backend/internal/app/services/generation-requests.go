package services

import (
	"errors"
	"math"

	"rip/internal/app/ds"
	"rip/internal/app/repositories"

	"gorm.io/gorm"
)

var (
	ErrorGenerationRequestIsNotDraft            = errors.New("generation request is not in \"draft\" status")
	ErrorGenerationRequestPeriodDaysRequired    = errors.New("period_days must be filled before generation request can be processed")
	ErrorGenerationRequestIncompleteTurbineData = errors.New("turbine data is incomplete for generation calculation")
	ErrorGenerationRequestDraftHasNoTurbines    = errors.New("generation request draft must contain at least one turbine")
	ErrorGenerationRequestCannotBeDeleted       = errors.New("only draft generation requests can be deleted")
	ErrorGenerationRequestIncorrectStatus       = errors.New("passed generation request status is incorrect")
)

type GenerationRequestsService struct {
	r *repositories.GenerationRequestRepository
}

func CalculateTurbineGeneration(avgVelocity float32, height uint16, alpha float32, power uint32, days uint) uint64 {
	const (
		optimalVelocity     = 12
		cutoffVelocity      = 15
		decreaseCoefficient = 0.4
	)

	velocityAtHeight := float64(avgVelocity) * math.Pow(float64(height)/10.0, float64(alpha))

	generation := float64(power) *
		math.Pow(velocityAtHeight/float64(optimalVelocity), 3) *
		math.Exp(-decreaseCoefficient*velocityAtHeight/float64(cutoffVelocity)) *
		24 * float64(days)

	return uint64(generation)
}

func NewGenerationRequestsService(db *gorm.DB) *GenerationRequestsService {
	return &GenerationRequestsService{
		r: repositories.NewGenerationRequestRepository(db),
	}
}

func (s *GenerationRequestsService) GetGenerationRequests(userId *uint, filter repositories.GenerationRequestsFilter) ([]ds.GenerationRequest, error) {
	return s.r.GetGenerationRequests(userId, filter)
}

func (s *GenerationRequestsService) GetGenerationRequest(generationRequestId uint) (ds.GenerationRequest, error) {
	return s.r.GetGenerationRequest(generationRequestId)
}

func (s *GenerationRequestsService) CloseGenerationRequest(generationRequestId uint, userId uint, status string) (ds.GenerationRequest, error) {
	switch status {
	case "completed":
		return s.r.CompleteGenerationRequest(generationRequestId, userId, CalculateTurbineGeneration)
	case "rejected":
		return s.r.RejectGenerationRequest(generationRequestId, userId)
	}

	return ds.GenerationRequest{}, ErrorGenerationRequestIncorrectStatus
}

func (s *GenerationRequestsService) GetDraftBriefInfo(userId uint) (ds.DraftGenerationRequestsBriefInfo, error) {
	currentDraft, err := s.r.GetDraftGenerationRequest(userId)
	if err != nil {
		return ds.DraftGenerationRequestsBriefInfo{}, err
	}

	currentDraftTurbinesCount, err := s.r.GetGenerationRequestsTurbinesCount(currentDraft.ID)
	if err != nil {
		return ds.DraftGenerationRequestsBriefInfo{}, err
	}

	return ds.DraftGenerationRequestsBriefInfo{
		GenerationRequestId: currentDraft.ID,
		TurbinesCount:       currentDraftTurbinesCount,
	}, nil
}

func (s *GenerationRequestsService) UpdateDraftGenerationRequest(userId uint, generationRequestUpdates ds.UpdateGenerationRequest) (ds.GenerationRequest, error) {
	return s.r.UpdateDraftGenerationRequest(userId, generationRequestUpdates)
}

func (s *GenerationRequestsService) AddTurbineToDraft(userId, turbineId uint) error {
	draftGenerationRequest, err := s.r.GetOrCreateDraftGenerationRequest(userId)
	if err != nil {
		return err
	}

	return s.r.AddTurbineToDraftGenerationRequest(draftGenerationRequest.ID, turbineId)
}

func (s *GenerationRequestsService) UpdateTurbineInDraft(userId, turbineId uint, updates ds.UpdateTurbineGenerationRequest) error {
	currentDraft, err := s.r.GetDraftGenerationRequest(userId)
	if err != nil {
		return err
	}

	return s.r.UpdateTurbineInDraft(currentDraft.ID, turbineId, updates)
}

func (s *GenerationRequestsService) RemoveTurbineFromDraft(userId, turbineId uint) error {
	currentDraft, err := s.r.GetDraftGenerationRequest(userId)
	if err != nil {
		return err
	}

	return s.r.RemoveTurbineFromDraftGenerationRequest(currentDraft.ID, turbineId)
}

func (s *GenerationRequestsService) SubmitDraftGenerationRequest(userId uint) error {
	return s.r.SubmitDraftGenerationRequest(userId)
}

func (s *GenerationRequestsService) DeleteDraftGenerationRequest(userId uint) error {
	currentDraft, err := s.r.GetDraftGenerationRequest(userId)
	if err != nil {
		return err
	}

	return s.r.DeleteGenerationRequest(currentDraft.ID)
}
