package client

import (
	"time"

	svc "github.com/sfs/pkg/service"
)

// structures containing the data fields needed for various
// pages in the client web interface

type Index struct {
	UserPage   string
	ProfilePic string
	UserName   string
	UserID     string
	Dirs       []*svc.Directory
	Files      []*svc.File
	ServerHost string
	ClientHost string
}

type FilePage struct {
	UserPage   string
	ProfilePic string
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
	UserPage     string
	ProfilePic   string
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
	UserPage   string
	ProfilePic string
	StatusCode string
	ErrMsg     string
	ServerHost string
	ClientHost string
}

type UserPage struct {
	UserPage       string
	ProfilePic     string
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
	UserPage     string
	ProfilePic   string
	UserID       string
	ServerHost   string
	ClientHost   string
	NoResultsMsg string
	Dirs         []*svc.Directory
	Files        []*svc.File
}

type SearchResults struct {
	Files []*svc.File      `json:"files"`
	Dirs  []*svc.Directory `json:"dirs"`
}

type AddPage struct {
	UserPage     string
	ProfilePic   string
	DiscoverPath string
	ServerHost   string
	ClientHost   string
	Endpoint     string
}

type EditPage struct {
	UserPage   string
	ProfilePic string
	ServerHost string
	ClientHost string
}

type UploadPage struct {
	UserPage   string
	ProfilePic string
	Dirs       []*svc.Directory
	ServerHost string
	ClientHost string
	Endpoint   string
}

type SettingsPage struct {
	UserPage   string
	ProfilePic string
	ServerHost string
	ClientHost string

	// Alterable settings
	UserName     string
	UserAlias    string
	UserEmail    string
	UserPassword string
	LocalSync    bool
}
