package mutexlimiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/config"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/infra/mutexlimiter"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Wait_Burst(t *testing.T) {
	token := bots.Token("token")
	rl := mutexlimiter.NewRateLimiter(config.RateLimiter{
		Capacity: 2,
		Rate:     0.1, // Не успеет восстановиться
	})

	wait, err := rl.Wait(context.Background(), token, timeWithSeconds(1))
	require.NoError(t, err)
	require.Zero(t, wait)

	wait, err = rl.Wait(context.Background(), token, timeWithSeconds(1))
	require.NoError(t, err)
	require.Zero(t, wait)

	wait, err = rl.Wait(context.Background(), token, timeWithSeconds(1))
	require.NoError(t, err)
	require.Equal(t, 10*time.Second, wait)
}

func TestRateLimiter_Wait_RateCheck(t *testing.T) {
	token := bots.Token("token")
	targetRate := 30.0
	rl := mutexlimiter.NewRateLimiter(config.RateLimiter{
		Capacity: 1,
		Rate:     targetRate,
	})

	n := 100
	i := n
	clock := timeWithSeconds(0)
	for i > 0 {
		wait, err := rl.Wait(context.Background(), token, clock)
		require.NoError(t, err)
		if wait > 0 {
			clock = clock.Add(wait)
		} else {
			i -= 1
		}
	}

	elapsed := clock.Sub(timeWithSeconds(0))
	rate := float64(n) / elapsed.Seconds()
	require.InDelta(t, targetRate, rate, 0.4)
}

func TestRateLimiter_Wait_MultipleTokens(t *testing.T) {
	rl := mutexlimiter.NewRateLimiter(config.RateLimiter{
		Capacity: 1,
		Rate:     1,
	})
	token1 := bots.Token("token1")
	token2 := bots.Token("token2")

	wait, err := rl.Wait(context.Background(), token1, timeWithSeconds(1))
	require.NoError(t, err)
	require.Zero(t, wait)

	wait, err = rl.Wait(context.Background(), token2, timeWithSeconds(1))
	require.NoError(t, err)
	require.Zero(t, wait)

	wait, err = rl.Wait(context.Background(), token1, timeWithSeconds(1))
	require.NoError(t, err)
	require.Equal(t, time.Second, wait)

	wait, err = rl.Wait(context.Background(), token2, timeWithSeconds(1))
	require.NoError(t, err)
	require.Equal(t, time.Second, wait)
}

func timeWithSeconds(s int) time.Time {
	return time.Date(2026, 4, 18, 0, 0, s, 0, time.UTC)
}
