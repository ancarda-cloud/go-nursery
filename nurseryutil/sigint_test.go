package nurseryutil_test

import (
	"os"
	"testing"
	"time"

	nursery "github.com/ancarda-cloud/go-lib-nursery"
	"github.com/ancarda-cloud/go-lib-nursery/nurseryutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // test relies on UNIX signals.
func TestShutdownOnInterrupt_ByClosingNursery(t *testing.T) {
	(nursery.Open(t.Context()))(func(nur nursery.Nursery) {
		err := nur.StartSoon(func(nur nursery.Nursery) {
			reason := nurseryutil.ShutdownOnInterrupt(nur, func() {
				t.Fatal("Was not expected to be called")
			})
			assert.Equal(t, nurseryutil.ShutdownReasonNurseryShutdown, reason)
		})
		require.NoError(t, err)
		nur.Shutdown()
	})
}

//nolint:paralleltest // test relies on UNIX signals.
func TestShutdownOnInterrupt_ByInterrupting(t *testing.T) {
	(nursery.Open(t.Context()))(func(nur nursery.Nursery) {
		err := nur.StartSoon(func(nur nursery.Nursery) {
			wasCalled := false
			reason := nurseryutil.ShutdownOnInterrupt(nur, func() {
				wasCalled = true
			})
			assert.Equal(t, nurseryutil.ShutdownReasonInterruptReceived, reason)
			assert.True(t, wasCalled)
		})
		require.NoError(t, err)
		interrupt(t)
	})
}

//nolint:paralleltest // test relies on UNIX signals.
func TestShutdownOnInterrupt_ByInterruptingWithNilHook(t *testing.T) {
	(nursery.Open(t.Context()))(func(nur nursery.Nursery) {
		err := nur.StartSoon(func(nur nursery.Nursery) {
			reason := nurseryutil.ShutdownOnInterrupt(nur, nil)
			assert.Equal(t, nurseryutil.ShutdownReasonInterruptReceived, reason)
		})
		require.NoError(t, err)
		interrupt(t)
	})
}

//nolint:paralleltest // test relies on UNIX signals.
func TestSetupShutdownOnInterrupt(t *testing.T) {
	(nursery.Open(t.Context()))(func(nur nursery.Nursery) {
		wasCalled := false
		err := nurseryutil.SetupShutdownOnInterrupt(nur, func() {
			wasCalled = true
		})
		require.NoError(t, err)
		interrupt(t)
		assert.Eventually(t, func() bool {
			return wasCalled
		}, time.Second, time.Millisecond)
	})
}

func TestSetupShutdownOnInterrupt_OnAlreadyShutdownNursery(t *testing.T) {
	t.Parallel()

	err := nurseryutil.SetupShutdownOnInterrupt(getShutdownNursery(t), func() {
		t.Fatal("Was not expected to be called")
	})

	require.Error(t, err)
	require.ErrorIs(t, err, nursery.ErrShuttingDown)
}

func interrupt(t *testing.T) {
	t.Helper()

	// Sometimes it takes just a tiny bit of time to start listening for the
	// signal, so this method sleeps for a tiny fraction of a second.
	time.Sleep(time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("cannot find test runner process: %s", err.Error())
	}

	err = proc.Signal(os.Interrupt)
	if err != nil {
		t.Fatalf("cannot raise interrupt: %s", err.Error())
	}
}

func getShutdownNursery(t *testing.T) nursery.Nursery {
	t.Helper()

	var shutdownNursery nursery.Nursery
	nursery.Open(t.Context())(func(nur nursery.Nursery) {
		shutdownNursery = nur
		shutdownNursery.Shutdown()
	})

	return shutdownNursery
}
