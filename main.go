package main

import (
	"log"

	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/server"
)

/*
  _____________________________
 /   _____/\_   _____/   _____/
 \_____  \  |    __) \_____  \
 /        \ |     \  /        \
/_______  / \___  / /_______  /
        \/      \/          \/
    (AKA) SIMPLE FILE SYNC
*/

// TODO:
// should deterime whether we're running client or server at runtime
func main() {
	if err := env.BuildEnv(true); err != nil {
		log.Fatal(err)
	}
	srv := server.NewServer()
	srv.Run()
}
