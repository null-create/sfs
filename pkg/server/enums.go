package server

type Context string

// context enums/keys for requests
const (
	Name        Context = "name"
	File        Context = "file"
	Files       Context = "files"
	Email       Context = "email"
	Admin       Context = "admin"
	Directory   Context = "directory"
	Directories Context = "directories"
	Parent      Context = "parent"
	Drive       Context = "drive"
	Path        Context = "path"
	User        Context = "user"
)
