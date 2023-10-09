package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

// this is mainly used for the one-time set up of the client-side service
func setup(userName, svcRoot string) (*Client, error) {
	svcPaths := []string{
		filepath.Join(svcRoot, "dbs"),
		filepath.Join(svcRoot, "state"),
	}

	// make sure these are empty before starting
	for _, svcPath := range svcPaths {
		entries, err := os.ReadDir(svcPath)
		if err != nil {
			return nil, err
		}
		if len(entries) != 0 {
			return nil, fmt.Errorf("svc path should be empty: %s", svcPath)
		}
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
