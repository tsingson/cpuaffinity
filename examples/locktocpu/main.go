package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/silentred/gid"

	"github.com/tsingson/cpuaffinity"
)

func randSleep() {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
}

func worker(id int, lock bool) {
	if lock {
		cpuaffinity.SetAffinity(id)
	}

	for {
		fmt.Println("main", gid.Get())
		fmt.Printf("worker: %d, CPU: %d\n", id, cpuaffinity.SchedGetCPU())
		randSleep()
	}
}

func main() {
	runtime.MemProfileRate = 0
	runtime.GOMAXPROCS(4)
	stopSignal := make(chan struct{})
	lock := len(os.Getenv("LOCK")) > 0
	lock = true
	// for i := 0; i < runtime.NumCPU(); i++ {
	cpuID := rand.Intn(runtime.NumCPU())
	go worker(cpuID, lock)
	// }
	<-stopSignal
}
