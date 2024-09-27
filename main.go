package main

import (
	"github.com/sfs/cmd"
	"github.com/sfs/pkg/configs"
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
	configs.SetEnv(false)
	cmd.Execute()
}
