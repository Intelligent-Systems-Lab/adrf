package app

// App defines the minimal lifecycle contract for ADRF service runtime.
//
// The interface is intentionally small so internal behavior can evolve without
// changing command entry and shutdown integration points.
type App interface {
	Start()
	Terminate()
	WaitRoutineStopped()
}
