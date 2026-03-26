package app

// App defines the minimal lifecycle contract for ADRF service runtime.
//
// The interface is intentionally small in R01 so later rounds can grow internal
// behavior without changing command entry and shutdown integration points.
type App interface {
	Start()
	Terminate()
	WaitRoutineStopped()
}
