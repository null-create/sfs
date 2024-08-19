package client

import (
	"time"

	svc "github.com/sfs/pkg/service"
)

// structures containing the data fields needed for various
// pages in the client web interface

type Index struct {
	UserName   string
	Dirs       []*svc.Directory
	Files      []*svc.File
	ServerHost string
	ClientHost string
}

type FilePage struct {
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
	StatusCode string
	ErrMsg     string
	ServerHost string
	ClientHost string
}

type UserPage struct {
	Name           string
	UserName       string
	Email          string
	TotalFiles     int
	TotalDirs      int
	ProfilePicPath string
	ServerHost     string
	ClientHost     string
}
