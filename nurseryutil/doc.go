// Package nurseryutil is a convenience package for common Nursery patterns.
//
// Functions starting with "Setup" start a nursery threads for you and return immediately,
// while functions without "Setup" block until they are done and should be started
// in a nursery thread you control using nur.StartSoon.
package nurseryutil
