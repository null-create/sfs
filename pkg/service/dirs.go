package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

	// client and server side paths
	ClientPath   string `json:"client_path"`
	ServerPath   string `json:"server_path"`
	BackupPath   string `json:"backup_path"`
	ServerBackup bool   `json:"server_backup"` // flag for whether this is the server-side version of the file
	LocalBackup  bool   `json:"local_backup"`  // flag for whether this is a local backup version of the file
	Registered   bool   `json:"registered"`    // flag for whether the file is registered with the server

	// security attributes
	Protected bool   `json:"protected"`
	AuthType  string `json:"auth_type"`
	Key       string `json:"-"`

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
	Parent   *Directory `json:"-"`
	ParentID string     `json:"parent_id"`

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
		Key:        auth.GenSecret(64),
		Overwrite:  false,
		LastSync:   time.Now().UTC(),
		Dirs:       make(map[string]*Directory, 0),
		Files:      make(map[string]*File, 0),
		Endpoint:   fmt.Sprint(Endpoint, ":", cfg.Port, "/v1/dirs/", uuid),
		Parent:     nil,
		ParentID:   "",
		Root:       true,
		Path:       rootPath,
		ServerPath: rootPath,
		ClientPath: rootPath,
		BackupPath: rootPath,
		RootPath:   rootPath,
		Registered: false,
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
		Key:        auth.GenSecret(64),
		Overwrite:  false,
		LastSync:   time.Now().UTC(),
		Dirs:       make(map[string]*Directory, 0),
		Files:      make(map[string]*File, 0),
		Endpoint:   fmt.Sprint(Endpoint, ":", cfg.Port, "/v1/dirs/", uuid),
		Parent:     nil,
		ParentID:   "",
		Root:       false,
		Path:       path,
		ClientPath: path,
		BackupPath: path,
		ServerPath: path,
		Registered: false,
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

// check if the physical directory actually exists
func (d *Directory) Exists() bool {
	if _, err := os.Stat(d.GetPath()); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// mark this instance as being a server-side back up of this directory
func (d *Directory) MarkServerBackup() {
	if !d.ServerBackup {
		d.ServerBackup = true
	}
}

// mark this instance as being a clinet-side backup of this directory
func (d *Directory) MarkLocalBackup() {
	if !d.LocalBackup {
		d.LocalBackup = true
	}
}

// retrieve the path to this directory depending whether its a
// server or client side directory
func (d *Directory) GetPath() string {
	var path string
	if d.ServerBackup {
		path = d.ServerPath
	} else {
		path = d.ClientPath
	}
	return path
}

// Remove all *internal data structure representations* of files and directories
// *Does not* remove actual files or sub directories themselves!
func (d *Directory) Clear() {
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
		d.Clear()
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
	err := filepath.Walk(d.GetPath(), func(filePath string, item os.FileInfo, err error) error {
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
	file.BackupPath = filepath.Join(d.BackupPath, file.Name)
	d.Size += file.GetSize()
	d.Files[file.ID] = file
	d.Files[file.ID].LastSync = time.Now().UTC()
	d.LastSync = time.Now().UTC()
}

// used when updating metadata for a file that's already in the directory.
// we don't need to modify file's directory info if this is the case.
func (d *Directory) putFile(file *File) {
	d.Files[file.ID] = file
	d.Files[file.ID].LastSync = time.Now().UTC()
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

// add a slice of files to this directory.
func (d *Directory) AddFiles(files []*File) {
	for _, f := range files {
		if err := d.AddFile(f); err != nil {
			log.Printf("failed to add file: %v", err)
		}
	}
}

// save new data to a file. file will be created or truncated,
// depending on its state at time of writing. does not check
// subdirectories for the existence of this file.
func (d *Directory) ModifyFile(file *File, data []byte) error {
	if !d.Protected {
		if d.HasFile(file.ID) {
			var origSize = file.GetSize()
			if err := file.Save(data); err != nil {
				return err
			}
			d.Size += file.GetSize() - origSize
		} else {
			var output = fmt.Sprintf(
				"file (id=%s) does not belong to this directory\nfile parent dir id=%s, cur dir id=%s\n",
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

// remove file from files map and update internal metadata.
// does not remove the physical file.
func (d *Directory) removeFile(fileID string) error {
	if file, ok := d.Files[fileID]; ok {
		delete(d.Files, file.ID)
		d.Size -= file.GetSize()
		d.LastSync = time.Now().UTC()
	} else {
		return fmt.Errorf("file (id=%s) not found", fileID)
	}
	return nil
}

// removes file from internal file map. does not remove the physical file.
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

// get a slice of all files from this directory, as well as its
// children.
// returns an empty slice if no files are found.
func (d *Directory) GetFiles() []*File {
	fileMap := d.WalkFs()
	if len(fileMap) == 0 {
		return make([]*File, 0)
	}
	files := make([]*File, 0, len(fileMap))
	for _, f := range fileMap {
		files = append(files, f)
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
		d.Dirs[subDir.ID] = subDir
	} else {
		return fmt.Errorf("dir (id=%s) not found. need to add before updating", subDir.ID)
	}
	return nil
}

// adds a new subdirectory and updates internal state.
// does *not* create a physical directory
func (d *Directory) addSubDir(dir *Directory) error {
	if _, exists := d.Dirs[dir.ID]; !exists {
		dir.Parent = d
		dir.ParentID = d.ID
		dir.DriveID = d.DriveID
		dir.BackupPath = filepath.Join(d.BackupPath, dir.Name)
		d.Dirs[dir.ID] = dir
		d.Dirs[dir.ID].LastSync = time.Now().UTC()
		d.LastSync = time.Now().UTC()
	} else {
		return fmt.Errorf("dir %s (id=%s) already exists", dir.Name, dir.ID)
	}
	return nil
}

// add a single sub directory to the current directory.
// does *not* create a physical directory!
func (d *Directory) AddSubDir(dir *Directory) error {
	if !d.Protected {
		if err := d.addSubDir(dir); err != nil {
			return err
		}
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
			if err := d.addSubDir(dir); err != nil {
				return err
			}
		}
	} else {
		log.Printf("%s (id=%s) is protected", d.Name, d.ID)
	}
	return nil
}

// remove from subdir map. returns nil if the directory is not found.
// does not remove physical directory!
func (d *Directory) removeDir(dirID string) *Directory {
	if sd, exists := d.Dirs[dirID]; exists {
		delete(d.Dirs, dirID)
		d.LastSync = time.Now().UTC()
		return sd
	}
	return nil
}

// removes subdirectory and *all of its child directories*
//
// returns a pointer to the subdirectory that was just removed, upon success.
// this should be useful for updating DBs after this operation completes.
func (d *Directory) RemoveSubDir(dirID string) *Directory {
	if !d.Protected {
		sd := d.removeDir(dirID)
		if sd == nil {
			log.Printf("directory (id=%s) not found", dirID)
			return nil
		}
		log.Printf("directory (id=%s) deleted", dirID)
		return sd
	} else {
		log.Printf("directory (id=%s) is protected", dirID)
		return nil
	}
}

// attempts to locate the directory or subdirectory starting from the given directory.
// returns nil if not found.
func (d *Directory) GetSubDir(dirID string) *Directory { return d.WalkD(dirID) }

// get a slice of all sub directories starting from the given directory.
// returns an empty slice if none are found.
func (d *Directory) GetSubDirs() []*Directory {
	dirMap := d.WalkDs()
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
func (d *Directory) GetDirMap() map[string]*Directory { return d.WalkDs() }

// recursively copies the directory tree to the given location.
//
// adapted from: https://stackoverflow.com/questions/51779243/copy-a-folder-in-go
func (d *Directory) CopyDir(src, dest string) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, PERMS); err != nil {
				return err
			}
			if err := d.CopyDir(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// ------------------------------------------------------------

/*
Walk() populates all files and subdirectory maps (and their files and subdirectories,
and so on) until we reach the end of the local directory tree.

Should be used only when instantiating a root directory object for the *first* time,
as it will generate new file and directory objects with their own ID's, and will need
to be treated as persistent items rather than ephemeral ones.
*/
func (d *Directory) Walk() *Directory {
	return walk(d)
}

// walk recursively descends the directory tree and populates all files
// and subdirectory maps in depth-first order.
func walk(d *Directory) *Directory {
	entries, err := os.ReadDir(d.GetPath())
	if err != nil {
		log.Printf("could not read directory: %v", err)
		return d
	}
	if len(entries) == 0 {
		return d
	}
	for _, entry := range entries {
		entryPath := filepath.Join(d.GetPath(), entry.Name())
		item, err := os.Stat(entryPath)
		if err != nil {
			log.Printf("could not get stat for %s - %v", entryPath, err)
			return d
		}
		if item.IsDir() {
			sd := NewDirectory(item.Name(), d.OwnerID, d.DriveID, entryPath)
			sd = walk(sd)
			if err := d.AddSubDir(sd); err != nil {
				log.Print(err)
			}
		} else {
			file := NewFile(item.Name(), d.DriveID, d.OwnerID, entryPath)
			if err := d.AddFile(file); err != nil {
				log.Print(err)
			}
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
	if len(dir.Files) == 0 {
		return nil
	}
	if file, found := dir.Files[fileID]; found {
		return file
	}
	for _, subDir := range dir.Dirs {
		if file := walkF(subDir, fileID); file != nil {
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
		files = walkFs(subDir, files)
	}
	return files
}

/*
WalkD() recursively traverses sub directories starting at a given directory (or root),
attempting to find the desired sub directory with the given directory ID.

Returns nil if the directory is not found
*/
func (d *Directory) WalkD(dirID string) *Directory {
	if d.ID == dirID {
		return d
	}
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
		dirMap = walkDs(subDir, dirMap)
	}
	return dirMap
}

func buildSync(dir *Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range dir.Files {
		if !idx.HasItem(file.ID) {
			idx.LastSync[file.ID] = file.LastSync
		}
	}
	// NOTE: monitoring directories is no longer supported.
	// if !idx.HasItem(dir.ID) {
	// 	idx.LastSync[dir.ID] = dir.LastSync
	// }
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
	for _, subDir := range dir.Dirs {
		idx = walkS(subDir, idx)
	}
	return idx
}

func buildUpdate(dir *Directory, idx *SyncIndex) *SyncIndex {
	for _, file := range dir.Files {
		if idx.HasItem(file.ID) {
			if file.LastSync.After(idx.LastSync[file.ID]) {
				idx.FilesToUpdate[file.ID] = file
			}
		}
	}
	// NOTE: monitoring Directories is no longer supported.
	// if idx.HasItem(dir.ID) {
	// 	if dir.LastSync.After(idx.LastSync[dir.ID]) {
	// 		idx.DirsToUpdate[dir.ID] = dir
	// 	}
	// }
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
	idx = buildUpdate(dir, idx)
	if len(dir.Dirs) == 0 {
		return idx
	}
	for _, subDir := range dir.Dirs {
		idx = walkU(subDir, idx)
	}
	return idx
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
