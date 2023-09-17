package auth

import (
	"encoding/json"
	"log"
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

	// pointer to the user's root nimbus drive
	// plus some meta data
	DriveID    string `json:"drive_id"`
	TotalFiles int    `json:"total_files"`
	TotalDirs  int    `json:"total_dirs"`
}

func check(name string, userName string, email string, newDrive string) bool {
	if name == "" || userName == "" || email == "" || newDrive == "" {
		return false
	}
	return true
}

func NewUser(name string, userName string, email string, newDrive string, isAdmin bool) *User {
	if !check(name, userName, email, newDrive) {
		log.Fatalf("[ERROR] all new user params must be provided")
	}
	return &User{
		ID:        NewUUID(),
		Name:      name,
		UserName:  userName,
		Password:  "default",
		Email:     email,
		LastLogin: time.Now(), // just to initalize the time.Time object

		Admin: isAdmin,

		DriveID: newDrive,
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
