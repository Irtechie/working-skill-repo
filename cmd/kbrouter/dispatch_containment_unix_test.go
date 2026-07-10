//go:build !windows

package main

func init() {
	// Cross-platform dispatcher fixtures use a fake executable and verify routing
	// logic, not production process containment. The dedicated Unix test restores
	// the production check and proves that a real worker cannot start.
	dispatchProcessTreeContainment = func() error { return nil }
}
