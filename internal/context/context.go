package context

import (
	"sync"
	"time"
)

var (
	self *ADRFContext
	once sync.Once
)

// ADRFContext holds process-wide runtime metadata.
//
// R01 intentionally keeps this structure small because ADRF state machines
// (store index, retrieval snapshots, fetch progression, retry counters) will be
// introduced in later rounds. A stable singleton is still created now so future
// state can be added without changing package boundaries.
type ADRFContext struct {
	AdrfName  string
	StartTime time.Time
}

// Init initializes the singleton runtime context exactly once.
//
// The singleton model matches free5gc NF style and prevents accidental divergent
// state copies across goroutines and modules.
func Init() {
	once.Do(func() {
		self = &ADRFContext{StartTime: time.Now().UTC()}
	})
}

// GetSelf returns the singleton ADRF runtime context.
func GetSelf() *ADRFContext {
	return self
}
