package client

import (
	"time"

	svc "github.com/sfs/pkg/service"
)

// structures containing the data fields needed for various
// pages in the client web interface

type Index struct {
	UserName   string
	Files      []*svc.File
	Dirs       []*svc.Directory
	ServerHost string
}

type FilePage struct {
	Name     string
	Size     int64
	Type     string
	Checksum string
	Endpoint string
	LastSync time.Time
}

type DirPage struct {
	Name         string
	Size         int64
	TotalFiles   int
	TotalSubDirs int
	Endpoint     string
	LastSync     time.Time
	SubDirs      []*svc.Directory
	Files        []*svc.File
}

type ErrorPage struct {
	StatusCode string
	ErrMsg     string
}

type UserPage struct {
	Name           string
	UserName       string
	Email          string
	TotalFiles     int
	TotalDirs      int
	ProfilePicPath string
}
