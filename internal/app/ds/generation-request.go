package ds

import (
	"encoding/json"
	"time"
)

type GenerationRequest struct {
	ID     uint   `json:"id"`
	Status string `gorm:"type:varchar(10);check:status IN ('draft','sent','completed','rejected','deleted');default:'draft';not null" json:"status"`

	CreatedByID    uint       `gorm:"not null" json:"-"`
	CreatedBy      User       `json:"-"`
	CreatedByLogin string     `json:"created_by"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
	FormedAt       *time.Time `json:"formed_at"`

	ClosedByID    *uint      `json:"-"`
	ClosedBy      *User      `json:"-"`
	ClosedByLogin *string    `json:"closed_by"`
	ClosedAt      *time.Time `json:"closed_at"`

	PeriodDays *uint `gorm:"check:(period_days > 0)" json:"period_days"`

	TurbineGenerationRequests      *[]TurbineGenerationRequest `json:"turbine_generation_requests,omitempty"`
	TurbineGenerationRequestsCount *uint                       `gorm:"-" json:"turbine_generation_requests_count"`
	CalculatedGenerationSum        *uint64                     `json:"calculated_generation_sum,omitempty"`
}


func (generationRequest GenerationRequest) MarshalJSON() ([]byte, error) {
	type Alias GenerationRequest

	var closedByLogin *string
	if generationRequest.ClosedBy != nil && generationRequest.ClosedBy.Login != "" {
		closedByLogin = &generationRequest.ClosedBy.Login
	}

	return json.Marshal(&struct {
		Alias
		CreatedByLogin string  `json:"created_by"`
		ClosedByLogin  *string `json:"closed_by"`
	}{
		Alias:          (Alias)(generationRequest),
		CreatedByLogin: generationRequest.CreatedBy.Login,
		ClosedByLogin:  closedByLogin,
	})
}

type UpdateGenerationRequest struct {
	PeriodDays *uint `json:"period_days" binding:"omitnil,gt=0,lt=36500"`
}

type DraftGenerationRequestsBriefInfo struct {
	GenerationRequestId uint
	TurbinesCount       uint
}
