package jwtauth

import (
	"github.com/golang-jwt/jwt/v5"
)

type accessTokenClaims struct {
	jwt.RegisteredClaims

	UserUUID string `json:"user_uuid"`
}

type AccessTokenPayload struct {
	UserUUID string
}
