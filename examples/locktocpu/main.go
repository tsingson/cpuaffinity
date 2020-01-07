package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/silentred/gid"

	"github.com/tsingson/cpuaffinity/affinity"
)

func randSleep() {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
}

func worker(id int, lock bool) {
	if lock {
		affinity.SetAffinity(id)
	}

	for {
		fmt.Println("main", gid.Get())
		fmt.Printf("worker: %d, CPU: %d\n", id, affinity.SchedGetCPU())
		randSleep()
	}
}

func main() {
	lock := len(os.Getenv("LOCK")) > 0
	lock = true
	// for i := 0; i < runtime.NumCPU(); i++ {
	cpuID := rand.Intn(runtime.NumCPU())
	go worker(cpuID, lock)
	// }
	time.Sleep(2 * time.Second)
}
