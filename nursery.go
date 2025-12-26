package nursery

import (
	"context"
	"fmt"
	"sync"
)

type nursery struct {
	ctx context.Context
	cf  context.CancelFunc
	wg  sync.WaitGroup
	mtx sync.Mutex
}

// Open creates a nursery and returns a callback function that can be used to
// begin executing tasks.
//
// This method blocks until the nursery is done executing.
func Open(ctx context.Context) func(CallbackFunc) {
	ctx, cf := context.WithCancel(ctx)

	nur := nursery{
		ctx: ctx,
		cf:  cf,
		wg:  sync.WaitGroup{},
		mtx: sync.Mutex{},
	}

	return nur.exec
}

func (nur *nursery) StartSoon(callback CallbackFunc) error {
	nur.mtx.Lock()
	defer nur.mtx.Unlock()

	err := nur.ctx.Err()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrShuttingDown, err)
	}

	nur.wg.Add(1)

	go func(nur *nursery, callback CallbackFunc) {
		defer nur.wg.Done()

		callback(nur)
	}(nur, callback)

	return nil
}

func (nur *nursery) Shutdown() {
	nur.mtx.Lock()
	defer nur.mtx.Unlock()

	nur.cf()
}

func (nur *nursery) Context() context.Context {
	return nur.ctx
}

func (nur *nursery) exec(callback CallbackFunc) {
	callback(nur)

	nur.wg.Wait()
	nur.cf()
}
