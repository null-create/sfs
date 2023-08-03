package pkg

import (
	"log"

	"github.com/nimbus/pkg/files"
)

type User struct {
	// user credentails
	ID        string        `json:"id"`
	NMap      files.NameMap `json:"nmap"`
	Name      string        `json:"name"`
	UserName  string        `json:"user_name"`
	Password  string        `json:"password"`
	Email     string        `json:"email"`
	LastLogin string        `json:"last_login"`

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
		ID:       files.NewUUID(),
		Name:     name,
		UserName: userName,
		Password: "default",
		Email:    email,

		Admin: isAdmin,

		Drive: newDrive,
	}
}
