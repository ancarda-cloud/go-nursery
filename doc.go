// Package nursery implements structured concurrency using the nursery pattern.
//
// To build and use a nursery, use the Open method, which returns a function
// you can invoke immediately; a CallbackFunc, which is called with a pointer
// to the nursery, and is able to launch goroutines this way:
//
//	do := nursery.Open(context.Background())
//	do(func(nur nursery.Nursery) {
//		_ = nur.Start(func(_ nursery.Nursery) {
//			log.Println("Hello")
//		})
//		_ = nur.Start(func(_ nursery.Nursery) {
//			log.Println("World")
//		})
//	})
//
// The above snippet will print "Hello World" or "World Hello", depending on
// random timing effects.
//
// You should not capture the Nursery object from the callback of Open. It will
// not be usable once shutdown, and it will shut down when your top level
// function returns.
package nursery
