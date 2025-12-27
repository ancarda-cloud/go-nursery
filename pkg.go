package nursery

import (
	"context"

	"git.sr.ht/~ancarda/consterr"
)

type (
	// CallbackFunc is a function that is called with a pointer to the nursery,
	// and is able to launch goroutines.
	CallbackFunc func(Nursery)

	// Nursery is an implementation of structured concurrency.
	//
	// All methods may be called from multiple goroutines simultaneously.
	Nursery interface {
		// Start spins up a goroutine and returns right away so the parent
		// function may call nursery functions.
		//
		// You will get ErrShuttingDown if the nursery is shutting down and
		// your goroutine was not launched. A non shutdown nursery will always
		// return nil.
		Start(nur CallbackFunc) error

		// Shutdown cancels the nursery, propagating cancellation to all
		// launched goroutines.
		//
		// After Shutdown has been called, all subsequent calls to Start
		// will fail with ErrShuttingDown.
		Shutdown()

		// Context returns the underlying context powering the nursery.
		// It is available so that you can pass it to services downstream
		// You can also select over Context().Done() if you wish to interrupt
		// goroutines to shut down faster.
		//
		// If Shutdown has been called, Context returns a canceled context.
		Context() context.Context
	}
)

// ErrShuttingDown is returned by time-consuming or concurrent functions when
// operations have been canceled or new operations have been refused.
const ErrShuttingDown consterr.E = "nursery is shutting down"
