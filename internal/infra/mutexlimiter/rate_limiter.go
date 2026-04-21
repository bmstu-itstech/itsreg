package mutexlimiter

import (
	"context"
	"sync"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/config"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

const eps = 1e-6

type bucket struct {
	tokens float64
	last   time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[bots.Token]*bucket
	capacity int
	rate     float64
}

func NewRateLimiter(cfg config.RateLimiter) *RateLimiter {
	return &RateLimiter{
		capacity: cfg.Capacity,
		rate:     cfg.Rate,
		buckets:  make(map[bots.Token]*bucket),
	}
}

func (r *RateLimiter) Wait(_ context.Context, token bots.Token, now time.Time) (time.Duration, error) {
	r.mu.Lock()
	b := r.bucket(token, now)

	// refill
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(float64(r.capacity), b.tokens+elapsed*r.rate)
	b.last = now

	// Наличие epsilon необходимо для того, чтобы учитывать погрешность при
	// вычислениях с плавающей точкой. Так, при отсутствии eps в тестах возникает
	// баг, связанный с тем, что b.tokens принимает значение 0.999... Это не проходит
	// условие if ниже, вследствие чего количество токенов не уменьшается. Но
	// вычисляя 1 - b.tokens происходит потеря точности (разность двух очень близких
	// чисел), и wait принимает значение нуля. Ситуация повторяется, и в итоге токены
	// не уменьшаются, а wait всегда равен нулю, что приводит к бесконечному циклу.
	// Баг воспроизводится только в тестах, так как используется идеальное время и
	// все запросы пришли одновременно.
	if b.tokens >= 1.0-eps {
		b.tokens -= 1.0
		r.mu.Unlock()
		return 0, nil
	}

	wait := time.Duration((1 - b.tokens) / r.rate * float64(time.Second))
	r.mu.Unlock()

	return wait, nil
}

func (r *RateLimiter) bucket(token bots.Token, now time.Time) *bucket {
	b, ok := r.buckets[token]
	if !ok {
		b = &bucket{
			tokens: float64(r.capacity),
			last:   now,
		}
		r.buckets[token] = b
	}
	return b
}
