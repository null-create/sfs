package client

import "github.com/sfs/pkg/service"

// structures for page data

type Index struct {
	UserName string
	Files    []*service.File
	Dirs     []*service.Directory
}

type FilePage struct {
	UserName string
	File     *service.File
}

type FolderPage struct {
	UserName string
	Dir      *service.Directory
}
