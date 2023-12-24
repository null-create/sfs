package main

import (
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
    		SIMPLE FILE SYNC
*/

// TODO:
// should deterime whether we're running client or server at runtime
func main() {
	env.SetEnv(true)
	srv := server.NewServer()
	srv.Run()
}
