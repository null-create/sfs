package server

type Context string

// context enums
const (
	Name      Context = "name"
	File      Context = "file"
	Files     Context = "files"
	Email     Context = "email"
	Admin     Context = "admin"
	Directory Context = "directory"
	Parent    Context = "parent"
	Drive     Context = "drive"
	Path      Context = "path"
	User      Context = "user"
)
