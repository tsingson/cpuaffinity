package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/silentred/gid"

	"github.com/tsingson/cpuaffinity/gofine"
)

func main() {
	var env gofine.Environment
	err := env.InitDefault()
	if err != nil {
		panic(err)
	}

	log.Printf("Available worker count: %v\n", env.LgoreCount())
	fmt.Println("main", gid.Get())
	var wg sync.WaitGroup
	wg.Add(1)
	go func(lgoreId int) {
		defer wg.Done()
		err := env.Occupy(lgoreId)
		if err != nil {
			panic(err)
		}
		defer env.Release(lgoreId)

		incrementMeHard := 0
		for {
			fmt.Println("main", gid.Get())
			// do non-interruptible super important work
			// open up htop and verify that goroutine doesn't jump around
			// and runs on the specified core index
			incrementMeHard++
		}
	}(0)

	wg.Wait()
}
