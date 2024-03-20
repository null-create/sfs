package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"

	"github.com/sfs/cmd"
	"github.com/sfs/pkg/env"
)

/*
  _____________________________
 /   _____/\_   _____/   _____/
 \_____  \  |    __) \_____  \
 /        \ |     \  /        \
/_______  / \___  / /_______  /
        \/      \/          \/
    		SIMPLE FILE SYNC
*/

func createPprofFiles() (*os.File, *os.File) {
	cpuFile, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	memFile, err := os.Create("mem.pprof")
	if err != nil {
		log.Fatal(err)
	}
	return cpuFile, memFile
}

func main() {
	// TEMP: for cpu and memory profiling
	// run go tool pprof -http=localhost:6060 sfs.exe cpu.pprof to see results
	cpuFile, memFile := createPprofFiles()

	// cpu profiling
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// memory profiling (heap allocation)
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		log.Fatal(err)
	}

	// main
	env.SetEnv(false)
	cmd.Execute()
}
