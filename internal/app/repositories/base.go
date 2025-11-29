package repositories

import (
	"errors"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrorNothingToUpdate = errors.New("got nothing to update")
)

func NewTurbinesDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
