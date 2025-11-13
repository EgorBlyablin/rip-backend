package services

import (
	"context"
	"errors"
	"fmt"
	"rip/internal/app/config"
	"rip/internal/app/ds"
	"rip/internal/app/redis"
	"rip/internal/app/repositories"
	"time"

	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrorUserCredentialsIncorrect = errors.New("user credentials are incorrect")
)

type UsersService struct {
	r                *repositories.UsersRepository
	Config           *config.Config
	Redis            *redis.Client
	JWTSigningMethod jwt.SigningMethod
}

func NewUsersService(db *gorm.DB, config *config.Config, redis *redis.Client) *UsersService {
	var jwtMethod jwt.SigningMethod
	switch config.JWT.SigningMethod {
	case "HS256":
		jwtMethod = jwt.SigningMethodHS256
	case "HS384":
		jwtMethod = jwt.SigningMethodHS384
	case "HS512":
		jwtMethod = jwt.SigningMethodHS512
	default:
		log.Fatalf("Unsupported JWT signing method: %s", config.JWT.SigningMethod)
	}

	return &UsersService{
		r:                repositories.NewUsersRepository(db),
		Config:           config,
		Redis:            redis,
		JWTSigningMethod: jwtMethod,
	}
}

func (s *UsersService) RegisterUser(user ds.CreateUser) (ds.User, error) {
	return s.r.CreateUser(user)
}

func (s *UsersService) GetUser(userId uint) (ds.User, error) {
	return s.r.GetUserByID(uint(userId))
}

func (s *UsersService) UpdateUser(userId uint, user ds.UpdateUser) (ds.User, error) {
	return s.r.UpdateUser(userId, user)
}

func (s *UsersService) Authorize(login string, password string) (string, error) {
	user, err := s.r.GetUserByLogin(login)
	if err != nil {
		return "", err
	}

	if user.Login == login && user.Password == password {
		token := jwt.NewWithClaims(s.JWTSigningMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(s.Config.JWT.ExpiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "turbines-backend",
			},
			UserID:      user.ID,
			IsModerator: user.IsModerator,
		})
		if token == nil {
			return "", fmt.Errorf("failed to create token")
		}

		signedToken, err := token.SignedString([]byte(s.Config.JWT.Token))
		if err != nil {
			return "", fmt.Errorf("can't create str token")
		}

		return signedToken, nil
	}

	return "", ErrorUserCredentialsIncorrect
}

func (s *UsersService) Deauthorize(ctx context.Context, jwtString string) error {
	_, err := jwt.ParseWithClaims(jwtString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Config.JWT.Token), nil
	})
	if err != nil {
		return err
	}

	err = s.Redis.WriteJWTToBlacklist(ctx, jwtString, s.Config.JWT.ExpiresIn)
	if err != nil {
		return err
	}

	return nil
}
