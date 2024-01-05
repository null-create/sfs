package pkg

import (
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
	env.SetEnv(false)

	// TODO: need way to shut one or both client and server down via command line
	// server stops with ctrl-c, but not the client
}
