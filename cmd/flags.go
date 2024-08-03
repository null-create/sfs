package cmd

type FlagPole struct {
	name string // file or directory name

	// ------- client cmd
	new      bool // create a new client
	start    bool // start a client
	local    bool // list all local files and directories managed by SFS
	remote   bool // list all remote files and directories managed by SFS
	refresh  bool // refresh local drive
	info     bool // get information about the client
	cleanBin bool // empty recycle bin

	// ----- drive command flags
	register  bool // register a new drive with the sfs server
	listFiles bool // list all files
	listDirs  bool // list all directories

	// discover command flags
	daemon bool // run in daemon mode

	// configs
	get     string
	setting string
	list    bool
	value   string

	// add and push cmd flags
	path     string
	isDir    bool
	newFile  bool
	newDir   bool
	discover bool // used to discover contents of entire file trees

	// copy cmd flags
	src  string
	dest string

	// ignore list flag
	ignore string

	// remove cmd
	delete bool // true to delete. false to just stop monitoring the item.

	autoSync  bool // flag to set auto sync mode
	localSync bool // flag to set local sync mode (local backups only)

	// remote server flags
	isUp  bool // flag to whether to see if the remote server is up
	stats bool // flag to check the stats of the remote server
}
