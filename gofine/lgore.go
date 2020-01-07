package gofine

import (
	"errors"
	"runtime"

	"golang.org/x/sys/unix"
)

// LgoreState represents lgore's current state
type LgoreState uint8

const (
	// Invalid state for non-existing lgores
	Invalid LgoreState = iota

	// Available represents an lgore which can be occupied
	Available

	// Busy represents an lgore which is occupied
	Busy
)

type lgore struct {
	coreId int
	state  LgoreState
}

func (lg *lgore) occupy() error {
	if lg.state == Busy {
		return errors.New("lgore is busy")
	}
	runtime.LockOSThread()

	var cpuset unix.CPUSet
	cpuset.Set(lg.coreId)

	err := unix.SchedSetaffinity(0, &cpuset)
	if err == nil {
		lg.state = Busy
	}
	return err
}

func (lg *lgore) release(original unix.CPUSet) error {
	if lg.state == Available {
		return nil
	}
	defer runtime.UnlockOSThread()

	lg.state = Available
	return unix.SchedSetaffinity(0, &original)
}
