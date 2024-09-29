package main

import (
	"log"
	_ "net/http/pprof"
	"runtime/pprof"

	"github.com/sfs/cmd"
	"github.com/sfs/pkg/configs"
	"github.com/sfs/pkg/utils"
)

/*
  _____________________________
 /   _____/\_   _____/   _____/
 \_____  \  |    __/ \_____  \
 /        \ |     \  /        \
/_______  / \___  / /_______  /
        \/      \/          \/
    		SIMPLE FILE SYNC
*/

func main() {
	// TEMP: for cpu and memory profiling
	// run: go tool pprof -http=localhost:6060 ./bin/sfs.exe <cpu.pprof or mem.pprof>
	// to see results
	cpuFile, memFile := utils.CreatePprofFiles()
	defer memFile.Close()
	defer cpuFile.Close()

	// cpu profiling
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// memory profiling (heap allocation)
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		log.Fatal(err)
	}

	configs.SetEnv(false)
	cmd.Execute()
}
