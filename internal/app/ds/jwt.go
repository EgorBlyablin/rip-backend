package ds

import (
	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims      // все что точно необходимо по RFC
	UserID             uint `json:"user_id"`
	IsModerator        bool `json:"is_moderator"`
}
