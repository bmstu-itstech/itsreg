package jwt_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/config"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/infra/jwt"
)

func TestTokenService(t *testing.T) {
	cfg := config.JWT{
		Secret:    "a-very-very-long-secret-string",
		AccessTTL: time.Second,
	}
	service, err := jwt.NewTokenService(cfg)
	require.NoError(t, err)
	uid := bots.UserID(1)

	token, err := service.GenerateToken(context.Background(), uid)
	require.NoError(t, err)

	parsedUID, err := service.VerifyToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, uid, parsedUID)

	require.Eventually(t, func() bool {
		_, err = service.VerifyToken(context.Background(), token)
		return errors.Is(err, port.ErrTokenInvalid) // Срок годности токена истёк
	}, 2*time.Second, time.Second)
}
