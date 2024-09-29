package utils

import (
	"log"
	"os"
)

func CreatePprofFiles() (*os.File, *os.File) {
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
