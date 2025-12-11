package ds

type TurbineGenerationRequest struct {
	TurbineID uint    `json:"turbine_id" gorm:"not null;uniqueIndex:idx_turbine_request"`
	Turbine   Turbine `json:"turbine"`

	GenerationRequestID uint              `json:"generation_request_id" gorm:"not null;uniqueIndex:idx_turbine_request"`
	GenerationRequest   GenerationRequest `json:"-"`

	AvgVelocity          *float32 `json:"avg_velocity" gorm:"type:numeric(4,1);check: avg_velocity > 0"`
	Alpha                *float32 `json:"alpha" gorm:"type:numeric(3,2);check: alpha > 0 AND alpha <= 1"`
	CalculatedGeneration *uint64  `json:"calculated_generation"`
}

type UpdateTurbineGenerationRequest struct {
	AvgVelocity          *float32 `json:"avg_velocity" binding:"omitnil,gt=0"`
	Alpha                *float32 `json:"alpha" binding:"omitnil,gt=0"`
	CalculatedGeneration *uint64  `json:"-"`
}
