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

func main() {
	// TEMP: for cpu profiling
	// run go tool pprof -http=localhost:6060 sfs.exe sfs.pprof to see results
	f, err := os.Create("sfs.pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// main
	env.SetEnv(false)
	cmd.Execute()
}
