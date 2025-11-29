package main

import (
	"rip/internal/app/ds"
	"rip/internal/app/dsn"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if err = db.AutoMigrate(
		&ds.User{},
		&ds.Turbine{},
		&ds.GenerationRequest{},
		&ds.TurbineGenerationRequest{},
	); err != nil {
		log.Error(err)
		panic("cant migrate db")
	}
}
