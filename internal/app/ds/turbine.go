package ds

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"rip/internal/app/config"
	"time"

	"github.com/gosimple/slug"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

const TurbinesBucket = "turbines"

type Turbine struct {
	ID          uint    `json:"id"`
	Title       string  `gorm:"type:varchar(100);not null" json:"title"`
	Description string  `gorm:"type:text;not null" json:"description"`
	IsActive    bool    `gorm:"default:1;not null" json:"is_active"`
	Image       *string `gorm:"type:varchar(100)" json:"Image"`

	Power  uint32 `gorm:"check:power > 0;not null" json:"power"`
	Height uint16 `gorm:"check:height > 0;not null" json:"height"`
}

var appConfig *config.Config

func (turbine *Turbine) ImageSrc() string {
	if turbine.Image == nil {
		return ""
	}

	if appConfig == nil {
		var err error
		appConfig, err = config.NewConfig()
		if err != nil {
			log.Fatalf("error loading config: %v", err)
		}
	}

	// Генерируем presigned URL на 1 час
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline") // чтобы браузер отображал, а не скачивал

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // доверяем самоподписанным сертификатам
		},
	}

	client, _ := minio.New(fmt.Sprintf("%s:%d", appConfig.Public.Host, appConfig.S3.Port), &minio.Options{
		Creds:     credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure:    true,
		Transport: transport,
	})
	urlStr, err := client.PresignedGetObject(
		context.Background(),
		TurbinesBucket,
		*turbine.Image,
		time.Hour, // срок жизни ссылки
		reqParams,
	)
	if err != nil {
		log.Errorf("error generating presigned URL for turbine image: %v", err)
		return ""
	}

	return urlStr.String()
}

func (turbine *Turbine) ImageSlug() string {
	return fmt.Sprintf("%d-%s", turbine.ID, slug.Make(turbine.Title))
}

func (turbine Turbine) MarshalJSON() ([]byte, error) {
	type Alias Turbine
	return json.Marshal(&struct {
		Alias
		Image string `json:"Image"`
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
	IsActive    *bool   `json:"is_active"`
	Image       *string

	Power  *uint32 `json:"power" binding:"omitnil,gt=0"`
	Height *uint16 `json:"height" binding:"omitnil,gt=0"`
}
