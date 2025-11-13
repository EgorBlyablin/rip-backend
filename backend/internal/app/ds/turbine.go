package ds

import (
	"encoding/json"
	"fmt"

	"github.com/gosimple/slug"
)

const TurbinesBucket = "turbines"

type Turbine struct {
	ID          uint    `json:"id"`
	Title       string  `gorm:"type:varchar(100);not null" json:"title"`
	Description string  `gorm:"type:text;not null" json:"description"`
	IsActive    bool    `gorm:"default:1;not null" json:"is_active"`
	Image       *string `gorm:"type:varchar(100)" json:"-"`

	Power  uint32 `gorm:"check:power > 0;not null" json:"power"`
	Height uint16 `gorm:"check:height > 0;not null" json:"height"`
}

func (turbine *Turbine) ImageSrc() string {
	if turbine.Image != nil {
		return fmt.Sprintf("http://localhost:9001/api/v1/buckets/%s/objects/download?preview=True&prefix=%s", TurbinesBucket, *turbine.Image)
	}
	return ""
}

func (turbine *Turbine) ImageSlug() string {
	return fmt.Sprintf("%d-%s", turbine.ID, slug.Make(turbine.Title))
}

func (turbine Turbine) MarshalJSON() ([]byte, error) {
	type Alias Turbine
	return json.Marshal(&struct {
		Alias
		Image string `json:"image"`
	}{
		Alias: (Alias)(turbine),
		Image: turbine.ImageSrc(),
	})
}

type CreateTurbine struct {
	Title       string `json:"title" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"required,min=10"`
	IsActive    bool   `json:"is_active" binding:"required"`

	Power  uint32 `json:"power" binding:"required,gt=0"`
	Height uint16 `json:"height" binding:"required,gt=0"`
}

type UpdateTurbine struct {
	Title       *string `json:"title" binding:"omitnil,min=3,max=100"`
	Description *string `json:"description" binding:"omitnil,min=10"`
	IsActive    *bool
	Image       *string

	Power  *uint32 `json:"power" binding:"omitnil,gt=0"`
	Height *uint16 `json:"height" binding:"omitnil,gt=0"`
}
