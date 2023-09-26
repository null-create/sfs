package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	// user credentails
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserName  string    `json:"user_name"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login"`

	// used for maintenance roles
	Admin bool `json:"admin"`

	// path to user state file
	SfPath string `json:"state_file"`

	// drive info
	DriveID    string `json:"drive_id"`
	TotalFiles int    `json:"total_files"`
	TotalDirs  int    `json:"total_dirs"`

	// path to the the drive root for their filesystem
	Root string `json:"root"`
}

func check(name, userName, email, newDrive, svcRoot string) bool {
	if name == "" || userName == "" || email == "" || newDrive == "" {
		return false
	}
	return true
}

func NewUser(name string, userName string, email string, newDriveID string, svcRoot string, isAdmin bool) *User {
	if !check(name, userName, email, newDriveID, svcRoot) {
		log.Fatalf("[ERROR] all new user params must be provided")
	}
	return &User{
		ID:        NewUUID(),
		Name:      name,
		UserName:  userName,
		Password:  "default",
		Email:     email,
		LastLogin: time.Now().UTC(),

		Admin: isAdmin,

		SfPath:  "", // set the first time the state is saved
		DriveID: newDriveID,
	}
}

// convert the curent user state to a json-formatted byte slice
func (u *User) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// save state to disk
func (u *User) SaveState() error {
	data, err := u.ToJSON()
	if err != nil {
		return err
	}
	fn := fmt.Sprintf("user-%s-.json", time.Now().UTC().Format("2006-01-02T15-04-05"))
	fp := filepath.Join(u.SfPath, fn)
	return os.WriteFile(fp, data, 0644)
}
