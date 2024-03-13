package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sfs/pkg/auth"
)

/*
File for handling all things related to creating the user's cloud directory

Should have a root directory called "nimbus" as a base line, after that, users
can specify the directory layout through a .yaml configuration, or through directory
creation at the command line.

NOTE:

	files contain the mutex lock, so lock time is dependent on whenever a method
	using a directory object calls a file's .Save() function within its internal Dirs list

	may want to explore moving lock use to the directory level, rather than the file level (?)
	or maybe just roll with it ...??
*/

type Directory struct {
	ID      string  `json:"id"`       // dir UUID
	NMap    NameMap `json:"nmap"`     // name map
	Name    string  `json:"name"`     // dir name
	OwnerID string  `json:"owner"`    // owner UUID
	DriveID string  `json:"drive_id"` // drive ID this directory belongs to

	// size in MB
	Size int64 `json:"size"`

	// absolute path to this directory.
	// should be something like:
	// .../sfs/user/root/../this_directory
	Path string `json:"path"`

	// directory client and server side paths
	ClientPath string `json:"client_path"`
	ServerPath string `json:"server_path"`

	// security attributes
	Protected bool   `json:"protected"`
	AuthType  string `json:"auth_type"`
	Key       string `json:"key"`

	// allows for automatic file overwriting or
	// directory replacement
	Overwrite bool `json:"overwrite"`

	// Last time this directory was modified
	LastSync time.Time `json:"last_sync"`

	// server API endpoint for this directory
	Endpoint string `json:"endpoint"`

	// map of files in this directory.
	// key is file uuid, value is file pointer
	Files map[string]*File `json:"-"`

	// map of subdirectories.
	// key is the directory uuid, value is the directory pointer
	Dirs map[string]*Directory `json:"-"`

	// pointer to parent directory (if not root).
	Parent *Directory

	// disignator for whether this directory is considerd the "root" directory
	Root     bool   `json:"root"`
	RootPath string `json:"root_path"`
}

// create a new root directory object. does not create physical directory.
func NewRootDirectory(dirName string, ownerID string, driveID string, rootPath string) *Directory {
	cfg := NewSvcCfg()
	uuid := auth.NewUUID()
	return &Directory{
		ID:         uuid,
		NMap:       newNameMap(dirName, uuid),
		Name:       dirName,
		OwnerID:    ownerID,
		DriveID:    driveID,
		Protected:  false,
		Key:        "default",
		Overwrite:  false,
		LastSync:   time.Now().UTC(),
		Dirs:       make(map[string]*Directory, 0),
		Files:      make(map[string]*File, 0),
		Endpoint:   fmt.Sprint(Endpoint, ":", cfg.Port, "/v1/dirs/", uuid),
		Parent:     nil,
		Root:       true,
		Path:       rootPath,
		ServerPath: rootPath,
		ClientPath: rootPath,
		RootPath:   rootPath,
	}
}

// NOTE: the Parent pointer isn't assigned by default! Must be assigned externally.
// This is mainly because I wanted an easier way to facilitate
// testing without having to create an entire mocked system.
// I'm sure I won't regret this.
func NewDirectory(dirName string, ownerID string, driveID string, path string) *Directory {
	cfg := NewSvcCfg()
	uuid := auth.NewUUID()
	return &Directory{
		ID:         uuid,
		NMap:       newNameMap(dirName, uuid),
		Name:       dirName,
		OwnerID:    ownerID,
		DriveID:    driveID,
		Protected:  false,
		Key:        "default",
		Overwrite:  false,
		LastSync:   time.Now().UTC(),
		Dirs:       make(map[string]*Directory, 0),
		Files:      make(map[string]*File, 0),
		Endpoint:   fmt.Sprint(Endpoint, ":", cfg.Port, "/v1/dirs/", uuid),
		Parent:     nil,
		Root:       false,
		Path:       path,
		ClientPath: path,
		ServerPath: path,
	}
}

// unmarshal a directory info string (usually retrieved from a request token)
// into a directory object.
func UnmarshalDirStr(data string) (*Directory, error) {
	dir := new(Directory)
	if err := json.Unmarshal([]byte(data), &dir); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dir data: %v", err)
	}
	return dir, nil
}

func (d *Directory) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Directory) IsRoot() bool { return d.Root }

func (d *Directory) HasParent() bool {
	return !d.IsRoot() && d.Parent == nil
}

func (d *Directory) IsNil() bool {
	return d.Files == nil && d.Dirs == nil
}

func (d *Directory) IsEmpty() bool {
	return len(d.Files) == 0 && len(d.Dirs) == 0
}

// check if the physical directory actually exists
func (d *Directory) Exists() bool {
	if _, err := os.Stat(d.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// returns the total count of files for only this directory.
func (d *Directory) TotalFiles() int { return len(d.Files) }

// return the total number of subdirectories for this directory
func (d *Directory) TotalSubDirs() int { return len(d.Dirs) }

// Remove all *internal data structure representations* of files and directories
// *Does not* remove actual files or sub directories themselves!
func (d *Directory) clear() {
	d.Files = nil
	d.Dirs = nil
	d.Files = make(map[string]*File, 0)
	d.Dirs = make(map[string]*Directory, 0)
}

// clean all files and subdirectories from the top-level directory.
// ***use with caution***
func clean(dirPath string) error {
	d, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("unable to read directory: %v", err)
	}
	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dirPath, name)); err != nil {
			return fmt.Errorf("unable to remove file or directory: %v", err)
		}
	}
	return nil
}

// calls clean() which ***removes all physical files and subdirectories
// from a drive starting at the given path***. use with caution!
//
// also calls dir.clear() which removes internal data structure references.
func (d *Directory) Clean(dirPath string) error {
	if !d.Protected {
		if err := clean(dirPath); err != nil {
			return err
		}
		d.clear()
		return nil
	} else {
		log.Printf("drive is protected.")
	}
	return nil
}

func (d *Directory) HasFile(fileID string) bool {
	if _, exists := d.Files[fileID]; exists {
		return true
	}
	return false
}

func (d *Directory) HasDir(dirID string) bool {
	if _, exists := d.Dirs[dirID]; exists {
		return true
	}
	return false
}

// returns the size of a directory with all its contents, including files in
// subdirectories. doesn't tally the size of subdirectories themselves, only counts
// file sizes.
func (d *Directory) GetSize() (int64, error) {
	var size int64
	err := filepath.Walk(d.Path, func(filePath string, item os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !item.IsDir() {
			size += item.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	d.Size = size
	return size, nil
}

/*
get the parent directory for this directory.

only root directories can have a nil parent pointer
since they will have a valid *drive pointer to
point to the parent drive
*/
func (d *Directory) GetParent() *Directory {
	if !d.HasParent() && !d.IsRoot() {
		log.Fatal("no parent for non-root directory!")
	}
	return d.Parent
}

// -------- password protection and other simple security stuff

func (d *Directory) SetPassword(password string, newPassword string) error {
	if password == d.Key {
		d.Key = newPassword
		log.Printf("password updated")
		return nil
	}
	return fmt.Errorf("wrong password")
}

func (d *Directory) Lock(password string) bool {
	if password == d.Key {
		d.Protected = true
		return true
	}
	log.Printf("wrong password")
	return false
}

func (d *Directory) Unlock(password string) bool {
	if password == d.Key {
		d.Protected = false
		return true
	}
	log.Printf("wrong password")
	return false
}

// --------- file management

// updates internal file map and file's sync time
func (d *Directory) addFile(file *File) {
	file.DirID = d.ID
	file.DriveID = d.DriveID
	file.LastSync = time.Now().UTC()
	d.Size += file.GetSize()
	d.Files[file.ID] = file
}

// used when updating metadata for a file that's already in the directory.
// we don't need to modify file's directory info if this is the case.
func (d *Directory) putFile(file *File) {
	file.LastSync = time.Now().UTC()
	d.Files[file.ID] = file
}

// add a file to this directory if not already present.
func (d *Directory) AddFile(file *File) error {
	if !d.Protected {
		if !d.HasFile(file.ID) {
			d.addFile(file)
		} else {
			return fmt.Errorf("file %s (id=%s) already present in directory", file.Name, file.ID)
		}
	} else {
		log.Printf("directory %s (id=%s) locked", d.Name, d.ID)
	}
	return nil
}

func (d *Directory) AddFiles(files []*File) {
	if !d.Protected {
		for _, f := range files {
			if !d.HasFile(f.ID) {
				d.addFile(f)
			} else {
				log.Printf("file (id=%s) already exists)", f.ID)
			}
		}
	} else {
		log.Printf("directory %s (id=%s) locked", d.Name, d.ID)
	}
}

// save new data to a file. file will be created or truncated,
// depending on its state at time of writing. does not check
// subdirectories for the existence of this file.
func (d *Directory) ModifyFile(file *File, data []byte) error {
	if !d.Protected {
		if d.HasFile(file.ID) {
			var origSize = file.GetSize()
			if err := os.WriteFile(file.ServerPath, data, PERMS); err != nil {
				return err
			}
			d.Size += file.GetSize() - origSize
		} else {
			var output = fmt.Sprintf(
				"file (id=%s) does not belong to this directory\nfile dirid=%s cur dirid=%s",
				file.ID, file.DirID, d.ID,
			)
			return fmt.Errorf(output)
		}
	} else {
		return fmt.Errorf("directory %s (id=%s) locked", d.Name, d.ID)
	}
	return nil
}

// update metadata for a file that's already in the directory.
func (d *Directory) PutFile(file *File) error {
	if !d.Protected {
		if d.HasFile(file.ID) {
			d.putFile(file)
		}
	} else {
		log.Printf("directory %s (id=%s) locked", d.Name, d.ID)
	}
	return nil
}

func (d *Directory) removeFile(fileID string) error {
	if file, ok := d.Files[fileID]; ok {
		if err := os.Remove(file.ServerPath); err != nil {
			return err
		}
		delete(d.Files, file.ID)
		d.Size -= file.GetSize()
		d.LastSync = time.Now().UTC()
	} else {
		return fmt.Errorf("file (id=%s) not found", file.ID)
	}
	return nil
}

// removes file from internal file map and deletes physical file.
// use with caution!
func (d *Directory) RemoveFile(fileID string) error {
	if !d.Protected {
		if err := d.removeFile(fileID); err != nil {
			return fmt.Errorf("failed to remove file: %s", err)
		}
	} else {
		log.Printf("directory protected. unlock before removing files")
	}
	return nil
}

// returns a file map containing all files starting at this directory.
func (d *Directory) GetFileMap() map[string]*File {
	return d.WalkFs()
}

// get a slice of all files starting from this directory.
// returns an empty slice if no files are found.
func (d *Directory) GetFiles() []*File {
	fileMap := d.WalkFs()
	if len(fileMap) == 0 {
		return make([]*File, 0)
	}
	var i int
	files := make([]*File, len(fileMap))
	for _, f := range fileMap {
		files[i] = f
		i++
	}
	return files
}

// find a file within the given directory or subdirectories, starting at the current directory.
// returns nil if no such file exists
func (d *Directory) FindFile(fileID string) *File {
	return d.WalkF(fileID)
}

// -------- sub directory methods

// update a subdirectory. must already exist as a subdirectory
// for this directory -- this is primarily used for updating a
// child subdirectory and reattaching it to its parent.
func (d *Directory) PutSubDir(subDir *Directory) error {
	if d.HasDir(subDir.ID) {
		subDir.Parent = d
		d.Dirs[subDir.ID] = subDir
	} else {
		return fmt.Errorf("dir (id=%s) not found. need to add before updating", subDir.ID)
	}
	return nil
}

// create a physical directory. does not check for whether
// this path is managed by the system, just creates a physical directory.
func (d *Directory) Mkdir(dirPath string) error {
	return os.Mkdir(dirPath, PERMS)
}

// adds a new subdirectory and updates internal state.
// does *not* create a physical directory
func (d *Directory) addSubDir(dir *Directory) error {
	if _, exists := d.Dirs[dir.ID]; !exists {
		dir.Parent = d
		dir.DriveID = d.DriveID
		d.Dirs[dir.ID] = dir
		d.Dirs[dir.ID].LastSync = time.Now().UTC()
	} else {
		return fmt.Errorf("dir %s (id=%s) already exists", dir.Name, dir.ID)
	}
	return nil
}

// add a single sub directory to the current directory.
// sets dir's parent pointer to the directory this
// function is attached to.
// does *not* create a physical directory!
// use dir.MkDir(path) instead.
func (d *Directory) AddSubDir(dir *Directory) error {
	if !d.Protected {
		d.addSubDir(dir)
	} else {
		log.Printf("dir %s is protected", d.Name)
	}
	return nil
}

// add a slice of directory objects to the current directory.
func (d *Directory) AddSubDirs(dirs []*Directory) error {
	if len(dirs) == 0 {
		return nil
	}
	if !d.Protected {
		for _, dir := range dirs {
			d.addSubDir(dir)
		}
	} else {
		log.Printf("%s (id=%s) is protected", d.Name, d.ID)
	}
	return nil
}

func (d *Directory) removeDir(dirID string) error {
	if dir, exists := d.Dirs[dirID]; exists {
		if err := os.RemoveAll(dir.Path); err != nil {
			return fmt.Errorf("unable to remove directory %s: %v", dirID, err)
		}
		delete(d.Dirs, dirID)
		log.Printf("directory (id=%s) removed", dirID)
	} else {
		log.Printf("directory (id=%s) is not found", dirID)
	}
	return nil
}

// removes a physical sub-directy and *all of its child directories*
// as well as the clearing the internal data structures.
//
// use with caution!
func (d *Directory) RemoveSubDir(dirID string) error {
	if !d.Protected {
		if err := d.removeDir(dirID); err != nil {
			return err
		}
		// remove from subdir map & update sync time
		delete(d.Dirs, dirID)
		d.LastSync = time.Now().UTC()
		log.Printf("directory (id=%s) deleted", dirID)
	} else {
		log.Printf("directory (id=%s) is protected", dirID)
	}
	return nil
}

// removes *ALL* sub directories and their children for a given directory
//
// calls d.Clean() which recursively deletes all subdirctories and their children
func (d *Directory) RemoveSubDirs() error {
	if !d.Protected {
		if err := d.Clean(d.Path); err != nil {
			return err
		}
		log.Printf("dir (id=%s) all sub directories deleted", d.ID)
	} else {
		log.Printf("dir (id=%s) is protected. no sub directories deleted", d.ID)
	}
	return nil
}

// attempts to locate the directory or subdirectory starting from the given directory.
// returns nil if not found.
func (d *Directory) GetSubDir(dirID string) *Directory {
	return d.WalkD(dirID)
}

// get a slice of all sub directories starting from the given directory.
// returns an empty slice if none are found.
func (d *Directory) GetSubDirs() []*Directory {
	dirMap := d.WalkDs()
	// convert to slice
	var i int
	dirs := make([]*Directory, len(dirMap))
	for _, d := range dirMap {
		dirs[i] = d
		i++
	}
	return dirs
}

// returns a map of all subdirectories starting from the current directory.
// returns an empty map if nothing is not found
func (d *Directory) GetDirMap() map[string]*Directory {
	return d.WalkDs()
}

// ------------------------------------------------------------

/*
Walk() populates all files and subdirectory maps (and their files and subdirectories,
and so on) until we reach the end of the local directory tree.
should be used only when instantiating a root directory object
for the *first* time, as it will generate new file and directory objects
with their own ID's, and will need to be treated as persistent items rather than
ephemeral ones.
*/
func (d *Directory) Walk() *Directory {
	if d.Path == "" {
		log.Print("can't traverse directory without a path")
		return d
	}
	return walk(d)
}

// walk recursively descends the directory tree and populates all files
// and subdirectory maps
func walk(d *Directory) *Directory {
	entries, err := os.ReadDir(d.Path)
	if err != nil {
		log.Printf("could not read directory: %v", err)
		return d
	}
	if len(entries) == 0 {
		return d
	}
	for _, entry := range entries {
		entryPath := filepath.Join(d.Path, entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			log.Printf("could not get stat for %s - %v", entryPath, err)
			return d
		}
		if item.IsDir() {
			sd := NewDirectory(item.Name(), d.OwnerID, d.DriveID, entryPath)
			sd = walk(sd)
			d.AddSubDir(sd)
		} else {
			file := NewFile(item.Name(), d.DriveID, d.OwnerID, entryPath)
			d.AddFile(file)
		}
	}
	return d
}

/*
WalkF() recursively traverses sub directories starting at a given directory (or root),
attempting to find the desired file with the given file ID.
*/
func (d *Directory) WalkF(fileID string) *File {
	return walkF(d, fileID)
}

func walkF(dir *Directory, fileID string) *File {
	if file, found := dir.Files[fileID]; found {
		return file
	}
	if len(dir.Dirs) == 0 {
		return nil
	}
	for _, subDirs := range dir.Dirs {
		if file := walkF(subDirs, fileID); file != nil {
			return file
		}
	}
	return nil
}

/*
WalkFs() recursively traversies all subdirectories and returns
a map of all files available for a given user
*/
func (d *Directory) WalkFs() map[string]*File {
	return walkFs(d, make(map[string]*File))
}

func walkFs(dir *Directory, files map[string]*File) map[string]*File {
	for _, file := range dir.Files {
		if _, exists := files[file.ID]; !exists {
			files[file.ID] = file
		}
	}
	if len(dir.Dirs) == 0 {
		return files
	}
	for _, subDir := range dir.Dirs {
		return walkFs(subDir, files)
	}
	return nil
}

/*
WalkD() recursively traverses sub directories starting at a given directory (or root),
attempting to find the desired sub directory with the given directory ID.

Returns nil if the directory is not found
*/
func (d *Directory) WalkD(dirID string) *Directory {
	return walkD(d, dirID)
}

func walkD(dir *Directory, dirID string) *Directory {
	if len(dir.Dirs) == 0 {
		return nil
	}
	if d, ok := dir.Dirs[dirID]; ok {
		return d
	}
	for _, subDirs := range dir.Dirs {
		if sd := walkD(subDirs, dirID); sd != nil {
			return sd
		}
	}
	return nil
}

/*
WalkDs() recursively traverses sub directories starting at a
given directory (or root) constructing a map of all sub directories.

Returns an empty map if nothing is not found
*/
func (d *Directory) WalkDs() map[string]*Directory {
	return walkDs(d, make(map[string]*Directory))
}

func walkDs(dir *Directory, dirMap map[string]*Directory) map[string]*Directory {
	if len(dir.Dirs) == 0 {
		return dirMap
	}
	for _, subDir := range dir.Dirs {
		if _, exists := dirMap[subDir.ID]; !exists {
			dirMap[subDir.ID] = subDir
		}
		return walkDs(subDir, dirMap)
	}
	return dirMap
}

func buildSync(dir *Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range dir.Files {
		if !idx.HasItem(file.ID) {
			idx.LastSync[file.ID] = file.LastSync
		}
	}
	// add directory's last sync time too
	if !idx.HasItem(dir.ID) {
		idx.LastSync[dir.ID] = dir.LastSync
	}
	return idx
}

/*
WalkS() recursively traverses each subdirectory starting from
the given directory and returns a *SyncIndex pointer containing
the last sync times for each file in each directory and subdirectories.
*/
func (d *Directory) WalkS(idx *SyncIndex) *SyncIndex {
	return walkS(d, idx)
}

func walkS(dir *Directory, idx *SyncIndex) *SyncIndex {
	idx = buildSync(dir, idx)
	if len(dir.Dirs) == 0 {
		return idx
	}
	for _, subDirs := range dir.Dirs {
		if sIndx := walkS(subDirs, idx); sIndx != nil {
			return sIndx
		}
	}
	return nil
}

func buildUpdate(dir *Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range dir.Files {
		if idx.HasItem(file.ID) {
			if file.LastSync.After(idx.LastSync[file.ID]) {
				idx.FilesToUpdate[file.ID] = file
			}
		}
	}
	if idx.HasItem(dir.ID) {
		if dir.LastSync.After(idx.LastSync[dir.ID]) {
			idx.DirsToUpdate[dir.ID] = dir
		}
	}
	return idx
}

// d.WalkU() populates the ToUpdate map of a given SyncIndex
func (d *Directory) WalkU(idx *SyncIndex) *SyncIndex {
	return walkU(d, idx)
}

// walkU recursively walks the directory tree and checks the last sync time
// of each file in each subdirectory, populating the ToUpdate map of a given SyncIndex
// as needed.
func walkU(dir *Directory, idx *SyncIndex) *SyncIndex {
	buildUpdate(dir, idx)
	if len(dir.Dirs) == 0 {
		return idx
	}
	for _, subDirs := range dir.Dirs {
		if uIndx := walkU(subDirs, idx); uIndx != nil {
			return uIndx
		}
	}
	return nil
}

// TODO: look into how to make this generic. this way
// we can loosen the requirements of the type for the argument
// to op(), and possibly the return type(s)

// WalkO() searches each subdirectory recursively and performes
// a supplied function on each file in the directory, returning
// an error if the function fails
//
// functions should have the following signature: func(file *File) error
func (d *Directory) WalkO(op func(file *File) error) error {
	return walkO(d, op)
}

func walkO(dir *Directory, op func(f *File) error) error {
	for _, file := range dir.Files {
		if err := op(file); err != nil {
			// we don't exit right away because this exception may only apply
			// to a single file.
			log.Printf("unable to run operation on %s \n%v\n continuing...", dir.Name, err)
			continue
		}
	}
	if len(dir.Dirs) == 0 {
		return nil
	}
	for _, subDirs := range dir.Dirs {
		if err := walkO(subDirs, op); err != nil {
			return err
		}
	}
	return nil
}
