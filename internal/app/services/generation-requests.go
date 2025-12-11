package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

type TurbineCalcRequest struct {
	TurbineID   uint    `json:"turbine_id"`
	AvgVelocity float32 `json:"avg_velocity"`
	Height      uint16  `json:"height"`
	Alpha       float32 `json:"alpha"`
	Power       uint32  `json:"power"`
	Days        uint    `json:"days"`
}

func sendCalculationRequest(generationRequestId uint, turbines []TurbineCalcRequest) error {
	body, _ := json.Marshal(turbines)
	url := fmt.Sprintf("http://async-backend:80/calculate-generation/%d", generationRequestId)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (s *GenerationRequestsService) CloseGenerationRequest(generationRequestId uint, userId uint, status string) (ds.GenerationRequest, error) {
	switch status {
	case "completed":
		generationRequest, err := s.r.CompleteGenerationRequest(generationRequestId, userId)

		if err == nil {
			turbinesCalcRequest := []TurbineCalcRequest{}

			for _, turbine := range *generationRequest.TurbineGenerationRequests {
				turbinesCalcRequest = append(turbinesCalcRequest, TurbineCalcRequest{
					TurbineID:   turbine.TurbineID,
					AvgVelocity: *turbine.AvgVelocity,
					Height:      turbine.Turbine.Height,
					Alpha:       *turbine.Alpha,
					Power:       turbine.Turbine.Power,
					Days:        *generationRequest.PeriodDays,
				})
			}

			sendCalculationRequest(generationRequestId, turbinesCalcRequest)
		}

		return generationRequest, err
	case "rejected":
		return s.r.RejectGenerationRequest(generationRequestId, userId)
	}

	return ds.GenerationRequest{}, ErrorGenerationRequestIncorrectStatus
}

func (s *GenerationRequestsService) UpdateTurbineInGenerationRequest(generationRequestId uint, turbineId uint, calculatedGeneration int) error {
	return s.r.UpdateTurbineInGenerationRequest(generationRequestId, turbineId, calculatedGeneration)
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
