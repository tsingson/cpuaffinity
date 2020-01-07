package gofine

import (
	"errors"
	"sync"

	"golang.org/x/sys/unix"
)

const maxNumCPUs = 1 << 10

var invalidLgoreId = errors.New("Invalid lgore id")

// Config environment config
type Config struct {
	// Specifies if we should pre-occupy all available cores
	//
	// TODO ignored for now
	OccupyAll bool

	// Specifies which cores should not be used as lgores
	//
	// There should be at least one core present in this slice.
	// Each value should be an index in range [0, NumCPUs),
	// where NumCPUs can be found from runtime.NumCPU().
	ReserveCores []int
}

// Environment manages all the lgores
type Environment struct {
	mu        sync.Mutex
	original  unix.CPUSet
	available unix.CPUSet
	lgores    []*lgore
}

// InitDefault initializes `env` with default configuration
//
// Default config sets `OccupyAll` to `false` and adds core 0 for reserve
func (env *Environment) InitDefault() error {
	defaultConf := Config{}
	defaultConf.OccupyAll = false
	defaultConf.ReserveCores = append(defaultConf.ReserveCores, 0)

	return env.Init(defaultConf)
}

// Init initializes environment and lgores
func (env *Environment) Init(conf Config) error {
	if len(conf.ReserveCores) == 0 {
		return errors.New("Should reserve at least one lgore")
	}

	// save original cpu affinity
	err := unix.SchedGetaffinity(0, &env.original)
	if err != nil {
		return err
	}

	if env.original.Count() <= 1 {
		return errors.New("Not enough logical cores, should be greater than one")
	}

	env.available = env.original

	// reserve cores for Go runtime
	for _, coreIndex := range conf.ReserveCores {
		coreId := getCoreIdByIndex(env.original, coreIndex)
		if coreId < 0 {
			return errors.New("Invalid reservation lgore id")
		}

		env.available.Clear(coreId)
	}
	if env.available.Count() == 0 {
		return errors.New("No lgores left after reservation")
	}

	env.initLgores()
	// TODO occupy lgores if OccupyAll is true
	return nil
}

// LgoreCount returns number of available lgores
func (env *Environment) LgoreCount() int {
	return env.available.Count()
}

// GetLgoreState returns `LgoreState` of a lgore
func (env *Environment) GetLgoreState(lgoreId int) (LgoreState, error) {
	if lgoreId >= len(env.lgores) {
		return Invalid, invalidLgoreId
	}

	return env.lgores[lgoreId].state, nil
}

// Occupy locks calling goroutine to an lgore
//
// Goroutine is locked to OS thread, and OS thread is locked to lgore's core.
func (env *Environment) Occupy(lgoreId int) error {
	if lgoreId >= len(env.lgores) {
		return invalidLgoreId
	}

	env.mu.Lock()
	defer env.mu.Unlock()

	lg := env.lgores[lgoreId]
	return lg.occupy()
}

// Release releases the lgore
//
// This function should be called from the same goroutine that called `Occupy`.
// Lgore becomes available, and the locked OS thread allowed to run on any core again.
func (env *Environment) Release(lgoreId int) error {
	if lgoreId >= len(env.lgores) {
		return invalidLgoreId
	}

	env.mu.Lock()
	defer env.mu.Unlock()

	lg := env.lgores[lgoreId]
	return lg.release(env.original)
}

func (env *Environment) initLgores() {
	env.lgores = make([]*lgore, env.available.Count())

	for lgoreId := 0; lgoreId < len(env.lgores); lgoreId++ {
		coreId := getCoreIdByIndex(env.available, lgoreId)
		env.lgores[lgoreId] = &lgore{coreId: coreId, state: Available}
	}
}

// returns -1 if not found
func getCoreIdByIndex(cpuset unix.CPUSet, coreIndex int) int {
	count := 0

	for coreId := 0; coreId < maxNumCPUs; coreId++ {
		if cpuset.IsSet(coreId) {
			if coreIndex == count {
				return coreId
			}

			count++
		}
	}

	return -1
}
