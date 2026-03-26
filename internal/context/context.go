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
// The structure is intentionally small. Additional runtime state can be added
// incrementally without changing package boundaries.
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
