package cmd

type FlagPole struct {
	name string // file or directory name

	// client cmd
	new     bool // create a new client
	start   bool // start a client
	local   bool // list all local files managed by SFS
	remote  bool // list all remote files managed by SFS
	refresh bool // refresh local drive
	info    bool // get information about the client

	// configs
	get bool
	set bool

	// add and push cmd flags
	path     string
	isDir    bool
	newFile  bool
	newDir   bool
	discover bool // used to discover contents of entire file trees

	// ignore list flag
	ignore string

	// remove cmd
	delete bool // true to delete. false to just stop monitoring the item.
}
