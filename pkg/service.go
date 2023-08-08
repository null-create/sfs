package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nimbus/pkg/files"
	"github.com/nimbus/pkg/server"
)

/*
Service Drive directory.
Container for all user Drives (dirs w/metadata and a sub dir for the users stuff).

Top level entry point for internal user file system and their operations.
Will likely be the entry point used for when a server is spun up.

All service configurations may end up living here.
*/
type Service struct {
	Name     string    `json:"name"`
	InitTime time.Time `json:"init_time"`

	// Drive directory path for Nimbus service on the server
	ServicePath string `json:"service_Drive"`

	// admin mode. allows for expanded permissions when working with
	// the internal nimbus file systems.
	AdminMode bool   `json:"admin_mode"`
	Admin     string `json:"admin"`
	AdminKey  string `json:"admin_key"`

	// key: drive-id, val is user struct.
	// user structs contain a pointer to the users Drive directory,
	// so this can be used for measuring disc size and executing
	// health checks
	Users map[string]*User `json:"users"`

	// HTTP server
	Srv *server.Server
}

// NOTE: http server is not instantiated with NewService()
func NewService(name string, admin bool) *Service {
	c := GetServiceConfig()
	svc := &Service{
		Name:        name,
		InitTime:    time.Now(),
		ServicePath: c.ServiceRoot,
		AdminMode:   admin,
		Users:       make(map[string]*User),
	}
	// input admin mode and credentials, if necessary
	if admin {
		s := server.SrvConfig()
		svc.AdminMode = true
		svc.Admin = s.Server.Admin
		svc.AdminKey = s.Server.AdminKey
	}
	return svc
}

func (s *Service) IsAdminMode() bool {
	return s.AdminMode
}

// returns the service run time in seconds
func (s *Service) RunTime() float64 {
	return time.Since(s.InitTime).Seconds()
}

// instantiate a new Nimbus server
func (s *Service) Start() error {
	s.Srv = server.NewServer()
	s.Srv.Start()

	return nil
}

// shut down service instance
func (s *Service) Stop() {
	s.Srv.Shutdown()
}

// get total active users
func (s *Service) TotalUsers() int {
	return len(s.Users)
}

func (s *Service) GetUsers() map[string]*User {
	if len(s.Users) == 0 {
		log.Printf("[DEBUG] no users found")
		return nil
	}
	return s.Users
}

func (s *Service) AddUser(u *User) {
	if s.Users == nil {
		s.Users = make(map[string]*User, 0)
	}
	s.Users[u.ID] = u
}

func (s *Service) RemoveUser(u *User) error {
	if usr, ok := s.Users[u.Drive.ID]; ok {
		delete(s.Users, u.Drive.ID)
	} else {
		return fmt.Errorf("[ERROR] user/drive %s not found", usr.Drive.ID)
	}
	return nil
}

// clear all active users drives and deletes all content within
func (s *Service) ClearAll(adminKey string) {
	if adminKey == s.AdminKey {
		if len(s.Users) == 0 {
			log.Printf("[DEBUG] no drives to remove")
			return
		}
		// remove all files and directories for this user
		log.Print("[DEBUG] cleaning...")
		for _, usr := range s.Users {
			usr.Drive.Root.Clean(usr.Drive.Root.RootPath)
			delete(s.Users, usr.Drive.ID)
			log.Printf("[DEBUG] user %s was removed", usr.UserName)
		}
	} else {
		log.Printf("[DEBUG] must enter admin password to clear all user drives")
	}
}

// get total size of all active user drives
func (s *Service) TotalSize() float64 {
	if len(s.Users) == 0 {
		log.Printf("[DEBUG] no drives to measure")
		return 0.0
	}
	var total float64
	for _, usr := range s.Users {
		total += usr.Drive.DriveSize()
	}
	return total
}

// Build a new privilaged Drive directory for a client on a Nimbus server
func (s *Service) AllocateDrive(name string, owner string) *files.Drive {
	drivePath := filepath.Join(s.ServicePath, name)

	newID := files.NewUUID()
	newRoot := files.NewRootDirectory(name, owner, filepath.Join(drivePath, name))
	drive := files.NewDrive(newID, name, owner, drivePath, newRoot)

	s.GenBaseFiles(drivePath)

	return drive
}

// TODO: test!
// generate some base line meta data for this service instance.
// should generate a users.json file (which will keep track of active users),
// and a drives.json, containing info about each drive, its total size, its location,
// owner, init date, passwords, etc.
func (s *Service) GenBaseFiles(DrivePath string) {
	// create Drive directory
	err := os.MkdirAll(DrivePath, 0666)
	if err != nil {
		log.Fatalf("[ERROR] failed to create Drive directory \n%v\n", err)
	}
	fileNames := []string{"user-info.json", "drive-info.json", "credentials.json"}
	for i := 0; i < len(fileNames); i++ {
		saveBaselineFile(DrivePath, fileNames[i], make(map[string]interface{}))
	}
}

// write out as a json file
func saveBaselineFile(dir, filename string, data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("[ERROR] failed marshalling JSON data: %s\n", err)
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", dir, filename), jsonData, 0644)
	if err != nil {
		log.Fatalf("[ERROR] unable to write JSON file %s: %s\n", filename, err)
		return
	}
}

// TODO:
// get the size of a hard drive. Will be useful for real-time health checks
func DiskSize(path string) (float64, error) {
	return 0.0, nil
}
