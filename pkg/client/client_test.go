package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/env"
	svc "github.com/sfs/pkg/service"

	"github.com/alecthomas/assert/v2"
)

// create a new client without a user
func TestNewClient(t *testing.T) {
	env.SetEnv(false)

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
	assert.NotEqual(t, nil, client.Monitor)
	assert.NotEqual(t, nil, client.Drive)
	assert.NotEqual(t, nil, client.Db)
	assert.NotEqual(t, nil, client.Handlers)
	assert.NotEqual(t, nil, client.Transfer)

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
	env.SetEnv(false)

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
	newUser, err := newUser(c1.Drive.Root.Path)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	newUser.DriveID = c1.Drive.ID
	c1.User = newUser
	if err = c1.SaveState(); err != nil {
		Fail(t, tmpDir, err)
	}

	// start a new client with this data and compare
	c2, err := Init(false)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// clean up before any possible assert failures
	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}

	assert.NotEqual(t, nil, c2)
	assert.Equal(t, c1.Conf, c2.Conf)
	assert.Equal(t, c1.User, c2.User)
	assert.Equal(t, c1.Root, c2.Root)
	assert.Equal(t, c1.User.ID, c2.User.ID)
	assert.Equal(t, c1.SfDir, c2.SfDir)
	assert.NotEqual(t, nil, c2.Db)
	assert.True(t, c2.Db.Singleton)
}

func TestLoadClientSaveState(t *testing.T) {
	env.SetEnv(false)

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

func TestLoadAndStartClient(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// initialize and load client
	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}
	if cSig, err := tmpClient.Start(); err == nil {
		log.Print("[TEST] adding test file...")
		time.Sleep(time.Millisecond * 500)

		// add a tmp file to see the client register the new file
		f, err := MakeTmpTxtFile(filepath.Join(tmpClient.Drive.Root.Path, "tmp.txt"), RandInt(1000))
		if err != nil {
			Fail(t, tmpDir, err)
		}

		// alter the file to see if monitor is detecting changes
		log.Print("[TEST] modifying test file...")
		MutateFile(t, f)
		time.Sleep(time.Millisecond * 500)

		// register the new file and apply more changes
		log.Print("[TEST] registering test file and altering again...")
		if err := tmpClient.AddFile(f.DirID, f); err != nil {
			Fail(t, tmpDir, err)
		}
		MutateFile(t, f)
		time.Sleep(time.Millisecond * 500)

		// stop client
		cSig <- os.Interrupt
	} else {
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

func TestClientUpdateUser(t *testing.T) {
	env.SetEnv(false)

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

	// update name initialized with client
	user := tmpClient.User
	user.Name = "William J Buttlicker"

	if err := tmpClient.UpdateUser(user); err != nil {
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
	env.SetEnv(false)

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

	user := tmpClient.User

	if err := tmpClient.RemoveUser(user.ID); err != nil {
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
	env.SetEnv(false)

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
	files := make([]*svc.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpClient.Root, fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpClient.Root, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	root := svc.NewDirectory("root", tmpClient.Conf.User, tmpClient.Drive.ID, tmpClient.Root)
	root.AddFiles(files)

	drv := svc.NewDrive(auth.NewUUID(), tmpClient.Conf.User, auth.NewUUID(), root.Path, root.ID, root)

	idx := drv.Root.WalkS(svc.NewSyncIndex(tmpClient.Conf.User))
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
// this test is flaky because we don't always alter
// files with each run. might need to not make it so random.
func TestClientBuildAndUpdateSyncIndex(t *testing.T) {
	env.SetEnv(false)

	// make sure we clean the right testing directory
	e := env.NewE()
	tmpDir, err := e.Get("CLIENT_ROOT")
	if err != nil {
		t.Fatal(err)
	}

	// make a test client
	tmpClient, err := Init(true)
	if err != nil {
		Fail(t, tmpDir, err)
	}

	// make a bunch of dummy files for this "client"
	total := RandInt(25)
	files := make([]*svc.File, 0, total)
	for i := 0; i < total; i++ {
		fn := filepath.Join(tmpClient.Root, fmt.Sprintf("tmp-%d.txt", i))
		if file, err := MakeTmpTxtFile(fn, RandInt(1000)); err == nil {
			files = append(files, file)
		} else {
			Fail(t, tmpClient.Root, err)
		}
	}

	// set up a new client drive and generate a last sync index of the files
	root := svc.NewDirectory("root", tmpClient.Conf.User, tmpClient.Drive.ID, tmpClient.Root)
	root.AddFiles(files)
	tmpClient.Drive = svc.NewDrive(auth.NewUUID(), tmpClient.Conf.User, tmpClient.UserID, root.Path, root.ID, root)

	// create initial sync index
	idx := tmpClient.Drive.Root.WalkS(svc.NewSyncIndex(tmpClient.Conf.User))

	// alter some files so we can mark them to be synced
	root.Files = MutateFiles(t, root.Files)

	// build ToUpdate map
	idx = tmpClient.Drive.Root.WalkU(idx)
	assert.NotEqual(t, nil, idx.ToUpdate)
	assert.NotEqual(t, 0, len(idx.ToUpdate))

	// clean up
	if err := Clean(t, tmpDir); err != nil {
		// reset our .env file for other tests
		if err2 := e.Set("CLIENT_NEW_SERVICE", "true"); err2 != nil {
			log.Fatal(err2)
		}
		log.Fatal(err)
	}
}
