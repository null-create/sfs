package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/service"

	"github.com/alecthomas/assert/v2"
)

// create a new client without a user
func TestNewClient(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client
	client, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, nil, client)
	assert.NotEqual(t, nil, client.Conf)
	assert.NotEqual(t, nil, client.User)
	assert.NotEqual(t, "", client.Root)
	assert.NotEqual(t, "", client.SfDir)
	// assert.NotEqual(t, nil, client.Monitor)
	assert.NotEqual(t, nil, client.Drive)
	assert.NotEqual(t, nil, client.Db)
	assert.NotEqual(t, nil, client.Handlers)
	// assert.NotEqual(t, nil, client.Transfer)

	// check that .env was updated after initialization,
	// specifically that CLIENT_NEW_SERVICE was set to "false"
	v, err := e.Get("CLIENT_NEW_SERVICE")
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, "false", v)

	// check for service directories and necessary databases
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		Fatal(t, err)
	}
	assert.NotEqual(t, 0, len(entries))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

// load a client with a pre-existing user
func TestLoadClient(t *testing.T) {
	env.BuildEnv(true)

	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	c1, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	// add a new user
	newUser, err := newUser("bill buttlicker", c1.Drive.ID, c1.Drive.Root.Path, e)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	c1.User = newUser
	if err = c1.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// start a new client with this data and compare
	c2, err := Init(false)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, nil, c2)
	assert.Equal(t, c1.Conf, c2.Conf)
	assert.Equal(t, c1.User, c2.User)
	assert.Equal(t, c1.Root, c2.Root)
	assert.Equal(t, c1.User.ID, c2.User.ID)
	assert.Equal(t, c1.SfDir, c2.SfDir)
	assert.Equal(t, c1.Db, c2.Db)
	assert.Equal(t, c1.Handlers, c2.Handlers)

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestLoadClientSaveState(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize a new client, then load a new client object
	// from the initialized service directory and state file
	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if err := tmpClient.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}
	entries, err := os.ReadDir(tmpClient.SfDir)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	assert.NotEqual(t, 0, len(entries))
	assert.Equal(t, 1, len(entries)) // should only have 1 state file at a time
	assert.True(t, strings.Contains(entries[0].Name(), "client-state"))
	assert.True(t, strings.Contains(entries[0].Name(), ".json"))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestClientUpdateUser(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	newUser := auth.NewUser(
		"bill buttlicker", "billB", "bill@bill.com", auth.NewUUID(),
		filepath.Join(tmpClient.Conf.Root, "bill buttlicker"), false,
	)

	if err := tmpClient.AddUser(newUser); err != nil {
		Fail(t, tmpDir, err)
	}

	newUser.Name = "william j buttlicker"
	if err := tmpClient.UpdateUser(newUser); err != nil {
		Fail(t, tmpDir, err)
	}

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestClientDeleteUser(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	newUser := auth.NewUser(
		"bill buttlicker", "billB", "bill@bill.com", auth.NewUUID(),
		filepath.Join(tmpClient.Conf.Root, "bill buttlicker"), false,
	)

	if err := tmpClient.AddUser(newUser); err != nil {
		Fail(t, tmpDir, err)
	}

	if err := tmpClient.RemoveUser(newUser.ID); err != nil {
		Fail(t, tmpDir, err)
	}
	assert.Equal(t, nil, tmpClient.User)

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestClientBuildSyncIndex(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for this "client"
	total := RandInt(25)
	files := make([]*service.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpClient.Root, fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpClient.Root, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	root := service.NewDirectory("root", tmpClient.Conf.User, tmpClient.Root)
	root.AddFiles(files)

	drv := service.NewDrive(auth.NewUUID(), tmpClient.Conf.User, tmpClient.Conf.User, root.Path, root)

	idx := drv.Root.WalkS(service.NewSyncIndex(tmpClient.Conf.User))
	assert.NotEqual(t, nil, idx)
	assert.NotEqual(t, 0, len(idx.LastSync))
	assert.Equal(t, total, len(idx.LastSync))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

// indexing and monitoring tests

func TestClientBuildAndUpdateSyncIndex(t *testing.T) {
	env.BuildEnv(true)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for this "client"
	total := RandInt(25)
	files := make([]*service.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpClient.Root, fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpClient.Root, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	root := service.NewDirectory("root", tmpClient.Conf.User, tmpClient.Root)
	root.AddFiles(files)
	drv := service.NewDrive(auth.NewUUID(), tmpClient.Conf.User, tmpClient.Conf.User, root.Path, root)
	tmpClient.Drive = drv

	idx := drv.Root.WalkS(service.NewSyncIndex(tmpClient.Conf.User))

	MutateFiles(t, root.Files)

	idx = drv.Root.WalkU(idx)
	assert.NotEqual(t, nil, idx.ToUpdate)
	assert.NotEqual(t, 0, len(idx.ToUpdate))

	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}

func TestClientContactServer(t *testing.T) {}

func TestClientGetSyncIndexFromServer(t *testing.T) {}

func TestClientSendSyncIndexToServer(t *testing.T) {}
