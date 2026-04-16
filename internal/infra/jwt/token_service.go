package jwt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/config"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type claims struct {
	jwt.RegisteredClaims

	UID int64 `json:"uid"`
}

type TokenService struct {
	cfg config.JWT
}

func NewTokenService(cfg config.JWT) (*TokenService, error) {
	if cfg.Secret == "" {
		return nil, errors.New("secret key required")
	}
	if cfg.AccessTTL == 0 {
		return nil, errors.New("access TTL required")
	}
	return &TokenService{cfg}, nil
}

func MustNewTokenService(cfg config.JWT) *TokenService {
	service, err := NewTokenService(cfg)
	if err != nil {
		panic(err)
	}
	return service
}

func (s *TokenService) GenerateToken(_ context.Context, userID bots.UserID) (string, error) {
	expiresAt := time.Now().Add(s.cfg.AccessTTL)

	c := &claims{
		UID: userID.Int64(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatInt(userID.Int64(), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *TokenService) VerifyToken(_ context.Context, token string) (bots.UserID, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("%w: %s", port.ErrTokenInvalid, err.Error())
	}

	c, ok := parsedToken.Claims.(*claims)
	if !ok || !parsedToken.Valid {
		return 0, port.ErrTokenInvalid
	}

	return bots.UserID(c.UID), nil
}
