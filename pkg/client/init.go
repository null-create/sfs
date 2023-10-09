package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

/*
define service directory paths, create necessary state file and
database directories, and create service databases and initial state file

NOTE: users files and directories within a dedicated service root.
"root"here means a dedicated directory for the user to backup and retrieve
any files and directories they wish.

A future alternative mode will to allow for individual files spread
across a user's normal system to be "marked" as files to "watch" for
activity (such as updates, modifications, etc), and then be queued for
synching or backing up with the server.

this can allow for more individual control over files and directories
as well as elmininate the need for a dedicated "root" service directory.
(not that this is an inherently bad idea, just want flexiblity)
*/
func setup(userName, svcRoot string) (*Client, error) {
	// make sure service root isn't already made
	if _, err := os.Stat(svcRoot); !os.IsNotExist(err) {
		return nil, fmt.Errorf("service root is already present: %v", err)
	}
	if err := os.Mkdir(svcRoot, svc.PERMS); err != nil {
		return nil, err
	}

	// define service directory paths
	svcPaths := []string{
		filepath.Join(svcRoot, "dbs"),
		filepath.Join(svcRoot, "root"),
		filepath.Join(svcRoot, "state"),
	}

	// make each directory
	for _, svcPath := range svcPaths {
		if err := os.Mkdir(svcPath, svc.PERMS); err != nil {
			return nil, err
		}
	}

	// make each database
	dbs := []string{"files", "directories"}
	for _, dName := range dbs {
		if err := db.NewDB(dName, svcPaths[0]); err != nil {
			return nil, err
		}
	}

	// set .env file CLIENT_NEW_SERVICE to false so we don't reinitialize every time
	e := svc.NewE()
	if err := e.Set("CLIENT_NEW_SERVICE", "false"); err != nil {
		return nil, err
	}

	return NewClient(userName, auth.NewUUID()), nil
}

func Setup() (*Client, error) {
	c := ClientConfig()
	client, err := setup(c.User, c.Root)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// initialize client service
func Init(newClient bool) (*Client, error) {
	if newClient {
		client, err := Setup()
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		client, err := LoadClient()
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}
