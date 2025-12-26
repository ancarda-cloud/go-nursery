package nurseryutil

import (
	"os"
	"os/signal"
	"syscall"

	nursery "github.com/ancarda-cloud/go-lib-nursery"
)

// ShutdownReason indicates why the nursery is being shutdown.
type ShutdownReason byte

const (
	// ShutdownReasonNurseryShutdown indicates a thread elsewhere called the
	// Shutdown() method on the nursery and the nursery is now shutting down as
	// a result.
	ShutdownReasonNurseryShutdown ShutdownReason = 0

	// ShutdownReasonInterruptReceived indicates that a SIGINT (^C) was received
	// and the nursery is now shutting down as a result.
	ShutdownReasonInterruptReceived ShutdownReason = 1
)

// ShutdownOnInterrupt listens for SIGINT (^C) and shuts down the nursery when
// that is received.
//
// The function blocks until the nursery is shut down, so you should wrap it in
// a nur.StartSoon call. Typically, you'd start this thread as soon as the
// nursery is created, e.g.:
//
//	do := nursery.Open(context.Background())
//	do(func(nur nursery.Nursery) {
//		_ = nur.StartSoon(func(nur nursery.Nursery) {
//			_ = ShutdownOnInterrupt(nur, nil)
//		})
//		// The rest of your code goes here...
//	})
//
// To perform actions as soon as SIGINT is raised, pass a function as the second
// argument. The nursery shutdown will be initiated after the method returns, so
// it's suggested you don't perform extensive work in the hook.
func ShutdownOnInterrupt(
	nur nursery.Nursery,
	onInterrupt func(),
) (reason ShutdownReason) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-nur.Context().Done():
		reason = ShutdownReasonNurseryShutdown
	case <-signalChan:
		reason = ShutdownReasonInterruptReceived
		if onInterrupt != nil {
			onInterrupt()
		}
	}

	nur.Shutdown()
	signal.Stop(signalChan)
	close(signalChan)

	return reason
}

// SetupShutdownOnInterrupt is a convenience wrapper around ShutdownOnInterrupt
// that starts the function in a new nursery thread and immediately returns.
//
// The shutdown reason is lost when using this function. If you need it, set up
// a thread yourself and call ShutdownOnInterrupt directly.
//
//	do := nursery.Open(context.Background())
//	do(func(nur nursery.Nursery) {
//		_ = nurseryutil.SetupShutdownOnInterrupt(nur, nil)
//		// The rest of your code goes here...
//	})
func SetupShutdownOnInterrupt(nur nursery.Nursery, onInterrupt func()) error {
	return nur.StartSoon(func(nur nursery.Nursery) {
		_ = ShutdownOnInterrupt(nur, onInterrupt)
	})
}
