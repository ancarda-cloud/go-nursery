package nursery_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ancarda-cloud/go-lib-nursery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveFunction(t *testing.T) {
	t.Parallel()

	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		assert.NoError(t, nur.Context().Err(), "Expected nursery to be active")
	})
}

func TestShutdown(t *testing.T) {
	t.Parallel()

	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		require.NoError(t, nur.Context().Err(), "Expected nursery to be active")
		nur.Shutdown()

		err := nur.Context().Err()
		require.Error(t, err, "Expected nursery to no longer be active")
		require.ErrorIs(t, err, context.Canceled)

		fatalFunc := func(_ nursery.Nursery) {
			panic("Expected nursery to be shutting down")
		}

		err = nur.StartSoon(fatalFunc)
		require.Error(t, err)
		require.ErrorIs(t, err, nursery.ErrShuttingDown)
	})
}

func TestGoroutinesWaitForNurseryToBeFinished(t *testing.T) {
	t.Parallel()

	ctr := atomic.Int32{}

	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		_ = nur.StartSoon(func(_ nursery.Nursery) {
			ctr.Add(1)
		})
		_ = nur.StartSoon(func(_ nursery.Nursery) {
			ctr.Add(1)
		})
	})

	assert.Equal(t, int32(2), ctr.Load(), "Expected two goroutines to have been executed")
}

func TestGoroutinesRunInParallel(t *testing.T) {
	t.Parallel()

	var startTime, endTime time.Time

	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		startTime = time.Now()

		for range 10_000 {
			_ = nur.StartSoon(func(_ nursery.Nursery) {
				time.Sleep(time.Millisecond)
			})
		}

		endTime = time.Now()
	})

	assert.Less(t, endTime.Sub(startTime), time.Millisecond*50)
}

func TestRetrieveContext(t *testing.T) {
	t.Parallel()

	type contextKey int

	ctx := context.WithValue(t.Context(), contextKey(1), "bar")

	nursery.Open(ctx)(func(nur nursery.Nursery) {
		val, _ := nur.Context().Value(contextKey(1)).(string)
		assert.Equal(t, "bar", val, "Expected to get back the same context")
	})
}

func TestCapturedNurseryRegistersAsShutdown(t *testing.T) {
	t.Parallel()

	var captured nursery.Nursery

	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		captured = nur
	})

	err := captured.Context().Err()
	require.Error(t, err, "Expected nursery to no longer be active")
	require.ErrorIs(t, err, context.Canceled)

	fatalFunc := func(_ nursery.Nursery) {
		panic("Expected nursery to be shutting down")
	}

	err = captured.StartSoon(fatalFunc)
	require.Error(t, err)
	assert.ErrorIs(t, err, nursery.ErrShuttingDown)
}

func BenchmarkNurseryOpen(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		nursery.Open(b.Context())
	}
}

func BenchmarkNurseryShutdown(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		nursery.Open(b.Context())(func(nur nursery.Nursery) {
			nur.Shutdown()
		})
	}
}

func BenchmarkNurseryStartSoon(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		nursery.Open(b.Context())(func(nur nursery.Nursery) {
			_ = nur.StartSoon(func(_ nursery.Nursery) {
			})
		})
	}
}
