package pkg

import (
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
	env.SetEnv(false)
	cmd.Execute()
}
