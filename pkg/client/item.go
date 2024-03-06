package client

import svc "github.com/sfs/pkg/service"

type Item struct {
	File      *svc.File
	Directory *svc.Directory
}

func (i *Item) HasFile() bool { return i.File != nil }
func (i *Item) HasDir() bool  { return i.Directory != nil }
