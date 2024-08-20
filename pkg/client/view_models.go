package client

import (
	"time"

	svc "github.com/sfs/pkg/service"
)

// structures containing the data fields needed for various
// pages in the client web interface

type Frame struct {
	UserID        string
	ProfilePicURL string
}

type Index struct {
	Frame      Frame
	UserName   string
	UserID     string
	Dirs       []*svc.Directory
	Files      []*svc.File
	ServerHost string
	ClientHost string
}

type FilePage struct {
	Frame      Frame
	Name       string
	Size       int64
	ID         string
	OwnerID    string
	Type       string
	LastSync   time.Time
	Location   string
	Checksum   string
	Endpoint   string
	ServerHost string
	ClientHost string
}

type DirPage struct {
	Frame        Frame
	Name         string
	Size         int64
	TotalFiles   int
	TotalSubDirs int
	Endpoint     string
	LastSync     time.Time
	Dirs         []*svc.Directory
	Files        []*svc.File
	ServerHost   string
	ClientHost   string
}

type ErrorPage struct {
	Frame      Frame
	StatusCode string
	ErrMsg     string
	ServerHost string
	ClientHost string
}

type UserPage struct {
	Frame          Frame
	Name           string
	UserID         string
	UserName       string
	Email          string
	TotalFiles     int
	TotalDirs      int
	ProfilePicPath string
	ServerHost     string
	ClientHost     string
}

type SearchPage struct {
	Frame      Frame
	UserID     string
	Dirs       []*svc.Directory
	Files      []*svc.File
	ServerHost string
	ClientHost string
}

type AddPage struct {
	Frame        Frame
	DiscoverPath string
	ServerHost   string
	ClientHost   string
}

type EditPageUpdates struct {
	Frame        Frame
	NewUserName  string
	NewUserAias  string
	NewUserEmail string
	ServerHost   string
	ClientHost   string
}

type UploadPage struct {
	Frame      Frame
	ServerHost string
	ClientHost string
}
