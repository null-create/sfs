package pkg

import (
	"log"
	"time"

	"github.com/sfs/pkg/files"
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

	// pointer to the user's root nimbus drive
	// plus some meta data
	Drive      *files.Drive
	TotalFiles int `json:"total_files"`
	TotalDirs  int `json:"total_dirs"`
}

func (u *User) IsAdmin() bool {
	return u.Admin
}

func check(name string, userName string, email string, newDrive *files.Drive) bool {
	if name == "" || userName == "" || email == "" || newDrive == nil {
		return false
	}
	return true
}

func NewUser(name string, userName string, email string, newDrive *files.Drive, isAdmin bool) *User {
	if !check(name, userName, email, newDrive) {
		log.Fatalf("[ERROR] all new user params must be provided")
	}
	return &User{
		ID:        files.NewUUID(),
		Name:      name,
		UserName:  userName,
		Password:  "default",
		Email:     email,
		LastLogin: time.Now(), // just to initalize the time.Time object

		Admin: isAdmin,

		Drive: newDrive,
	}
}
