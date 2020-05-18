package service

import (
	"sync/atomic"
	"time"
)

// atomicNow is the current time in minute precision.
var atomicNow atomic.Value

// now returns the current time value stored in atomicNow.
func now() time.Time { return atomicNow.Load().(time.Time) }

func init() {
	// start an update loop for atomicNow.
	go func() {
		for {
			atomicNow.Store(time.Now())
			time.Sleep(time.Minute)
		}
	}()
}
