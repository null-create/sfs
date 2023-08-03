package files

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
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
	ID    string  `json:"id"`
	NMap  NameMap `json:"name_map"`
	Name  string  `json:"name"`
	Owner string  `json:"owner"`

	// size in MB
	Size float64 `json:"size"`

	// absolute path to this directory.
	// should be something like:
	// .../nimbus/user/root/../this_directory
	Path string `json:"path"`

	Protected bool   `json:"protected"` //
	AuthType  string `json:"auth_type"`
	Key       string `json:"key"`

	// allows for automatic file overwriting or
	// directory replacement
	Overwrite bool `json:"overwrite"`

	// Last time this directory was modified
	LastSync time.Time `json:"last_sync"`

	// key is file uuid, value is file pointer
	Files map[string]*File

	// map of subdirectories.
	// key is the directory ID (UUID), value is the directory pointer
	Dirs map[string]*Directory

	// pointer to parent directory (if not root).
	Parent *Directory

	// disignator for whether this directory is considerd the "root" directory
	// if Root is True, then Drive will not be nil (Parent will be nil!) and
	// should point to the parent Drive struct
	Root     bool   `json:"root,omitempty"`
	RootPath string `json:"rootPath,omitempty"`
}

func NewRootDirectory(name string, owner string, rootPath string) *Directory {
	uuid := NewUUID()
	return &Directory{
		ID:        uuid,
		NMap:      newNameMap(name, uuid),
		Name:      name,
		Owner:     owner,
		Protected: false,
		Key:       "default",
		Overwrite: false,
		Dirs:      make(map[string]*Directory, 0),
		Files:     make(map[string]*File, 0),
		Parent:    nil,
		Root:      true,
		Path:      rootPath,
		RootPath:  rootPath,
	}
}

// NOTE: the Parent pointer isn't assigned by default! Must be assigned externally.
// This is mainly because I wanted an easier way to facilitate
// testing without having to create an entire mocked system.
// I'm sure I won't regret this.
func NewDirectory(name string, owner string, path string) *Directory {
	uuid := NewUUID()
	return &Directory{
		ID:        uuid,
		NMap:      newNameMap(name, uuid),
		Name:      name,
		Owner:     owner,
		Protected: false,
		Key:       "default",
		Overwrite: false,
		Dirs:      make(map[string]*Directory, 0),
		Files:     make(map[string]*File, 0),
		Root:      false,
		Path:      path,
	}
}

func (d *Directory) IsRoot() bool {
	return d.Root
}

func (d *Directory) IsProtected() bool {
	return d.Protected
}

func (d *Directory) HasParent() bool {
	if d.Parent == nil {
		log.Print("[ERROR] parent directory cannot be nil")
		return false
	}
	return true
}

// Remove all *internal data structure representations* of files and directories
// *Does not* remove actual files or sub directories themselves!
func (d *Directory) clear() {
	d.Files = make(map[string]*File, 0)
	d.Dirs = make(map[string]*Directory, 0)

	log.Printf("[DEBUG] dirID(%s) all directories and files deleted", d.ID)
}

// used to securely run clear()
func (d *Directory) Clear(password string) {
	if !d.IsProtected() {
		d.clear()
	} else {
		log.Print("[DEBUG] directory protected")
		if password == d.Key {
			d.clear()
		} else {
			log.Print("[DEBUG] wrong password. contents not deleted")
		}
	}
}

/*
recursively cleans all contents from a directory and its subdirectories

***VERY DANGEROUS AND SHOULD ONLY BE USED IN A SECURE CONTEXT***

mainly just want to make sure dirPath is what we *actually* want to clean,
not a system/OS path or anything vital of any kind :(
*/
func (d *Directory) clean(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// keep going if we find a subdirectory
			if err = d.clean(path); err != nil {
				return err
			}
			// remove the directory
			if err = os.Remove(path); err != nil {
				return err
			}
		} else {
			// remove the file
			if err = os.Remove(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// calls d.clean(), which recursively removes all files and
// subdirectories from a drive
func (d *Directory) Clean(dirPath string) error {
	if !d.IsProtected() {
		d.clean(dirPath)
		return nil
	} else {
		log.Printf("[DEBUG] drive is protected.")
	}
	return nil
}

// TODO: look into more efficient ways to check for
// existence of keys in map[string]Type objects
// d.Files and d.Directories are both map[string]Type respectively,
// so we should theoretically be able to just try a key and see if it
// works, rather than iterate over every item since the keys in both
// maps are the files and directories UUIDs.
// O(n) vs O(1)!

/*
	if err, ok := d.Directories[fileID]; ok {
		return true
	}
*/

func (d *Directory) HasFile(fileID string) bool {

	for _, f := range d.Files {
		if f.ID == fileID {
			return true
		}
	}

	return false
}

func (d *Directory) HasDir(dirID string) bool {
	for _, dir := range d.Dirs {
		if dir.ID == dirID {
			return true
		}
	}
	return false
}

/*
only root directories can have a nil parent pointer
since they will have a valid *drive pointer to
point to the parent drive
*/
func (d *Directory) GetParent() *Directory {
	if d.Parent == nil && !d.Root {
		log.Fatal("[ERROR] no parent for directory")
		return nil
	} else {
		return d.Parent
	}
}

// -------- password protection and other simple security stuff

func (d *Directory) SetPassword(password string, newPassword string) error {
	if password == d.Key {
		d.Key = newPassword
		log.Printf("[DEBUG] password updated")
		return nil
	}
	return fmt.Errorf("[ERROR] wrong password")
}

// protected OPs
func (d *Directory) Lock(password string) bool {
	if password == d.Key {
		d.Protected = true
		return true
	}
	log.Printf("[DEBUG] wrong password")
	return false
}

func (d *Directory) Unlock(password string) bool {
	if password == d.Key {
		d.Protected = false
		return true
	}
	log.Printf("[DEBUG] wrong password")
	return false
}

// --------- file management

func (d *Directory) addFile(f *File) {
	d.Files[f.ID] = f
	d.Files[f.ID].LastSync = time.Now()

	// load into memory just to create the file,
	if len(f.Content) == 0 {
		f.Load()
	}
	f.Save(f.Content)
	f.Clear()

	log.Printf("[DEBUG] file %s (%s) added", f.Name, f.ID)
}

func (d *Directory) AddFile(file *File) {
	if !d.HasFile(file.ID) {
		d.addFile(file)
	} else {
		log.Printf("[DEBUG] file %s (%s) already present in directory", file.Name, file.ID)
	}
}

func (d *Directory) AddFiles(files []*File) {
	if len(files) == 0 {
		log.Printf("[DEBUG] no files recieved")
		return
	}
	for _, f := range files {
		if !d.HasFile(f.ID) {
			d.addFile(f)
		} else {
			log.Printf("[DEBUG] file %v already exists)", f.ID)
		}
	}
}

func (d *Directory) removeFile(fileID string) error {
	removed := false
	for _, f := range d.Files {
		if f.ID == fileID {
			// remove actual file
			err := os.Remove(f.ServerPath)
			if err != nil {
				return fmt.Errorf("[ERROR] unable to remove file: %s \n%v\n ", f.Name, err)
			}

			// update internal files list and directory's last sync time
			delete(d.Files, f.ID)
			d.LastSync = time.Now()
			removed = true

			log.Printf("[DEBUG] file %s removed", f.ID)
		}
	}
	if !removed {
		log.Printf("[DEBUG] no file with ID %v was found or removed", fileID)
	}
	return nil
}

// Removes actual file plus internal File object
func (d *Directory) RemoveFile(fileID string) error {
	if !d.IsProtected() {
		d.removeFile(fileID)
	} else {
		log.Printf("[DEBUG] directory protected. unlock before removing files")
	}
	return nil
}

// returns a slice of NameMap objects representing files and their UUIDs
// from a given directory. does not return info about files in any subdirectories
func (d *Directory) GetFileList() []NameMap {
	if len(d.Files) == 0 {
		log.Print("[DEBUG] file list is empty")
		return nil
	}
	fileIDs := make([]NameMap, len(d.Files))
	for _, f := range d.Files {
		fileIDs = append(fileIDs, newNameMap(f.Name, f.ID))
	}
	return fileIDs
}

// -------- sub directory methods

func (d *Directory) addSubDir(dir *Directory) {
	dir.Parent = d
	d.Dirs[dir.ID] = dir
	d.LastSync = time.Now()

	if err := os.MkdirAll(dir.Path, PERMS); err != nil {
		log.Fatalf("[ERROR] could not create directory")
	}

	log.Printf("[DEBUG] dir %s (%s) added", dir.Name, dir.ID)
}

// add a single sub directory to the current directory
func (d *Directory) AddSubDir(dir *Directory) {
	if !d.IsProtected() {
		d.addSubDir(dir)
	} else {
		log.Printf("[DEBUG] dir %s is protected", d.Name)
	}
}

// add a slice of directory objects to the current directory
func (d *Directory) AddSubDirs(dirs []*Directory) {
	if len(dirs) == 0 {
		log.Printf("[DEBUG] dir list is empty")
		return
	}
	if !d.IsProtected() {
		for _, dir := range dirs {
			d.addSubDir(dir)
		}
	} else {
		log.Printf("[DEBUG] directory (id=%s) is protected", d.ID)
	}
}

// removes a subdirecty and *all of its child directories*
// use with caution!
func (d *Directory) RemoveSubDir(dir *Directory) {
	if d.HasDir(dir.ID) {
		delete(d.Dirs, dir.ID) // remove from d.Dirs
		d.Clean(dir.ID)        // remove actual subdir and all its children
		d.LastSync = time.Now()

		log.Printf("[DEBUG] directory %s (%s) deleted", dir.Name, dir.ID)
	} else {
		log.Printf("[DEBUG] directory %s (%s) not found", dir.Name, dir.ID)
	}
}

// removes *ALL* sub directories and their children for a given directory
func (d *Directory) RemoveSubDirs() {
	if !d.IsProtected() {
		d.Dirs = make(map[string]*Directory, 0)
		d.Clean(d.Path)

		log.Printf("[DEBUG] dir(%s) all sub directories deleted", d.ID)
	} else {
		log.Printf("[DEBUG] dir(%s) is protected. no sub directories deleted", d.ID)
	}
}

// directly returns the subdirectory, assuming the supplied key is valid
func (d *Directory) GetSubDir(dirID string) *Directory {
	if dir, ok := d.Dirs[dirID]; ok {
		return dir
	} else {
		log.Printf("[DEBUG] dir(%s) not found", dirID)
		return nil
	}
}

func (d *Directory) GetSubDirs() map[string]*Directory {
	if len(d.Dirs) == 0 {
		log.Print("[DEBUG] sub directory list is empty")
		return nil
	}
	return d.Dirs
}

// Returns the size of a directory with all its contents.
func (d *Directory) DirSize() (float64, error) {
	var size float64

	err := filepath.Walk(d.Path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// TODO: investigate how the conversion of int64 to float64 can effect results.
			size += float64(info.Size())
		}
		return nil
	})

	return size, err
}

/*
Walk functions:



*/

/*
Walk() recursively traverses sub directories starting at a given directory (or root),
attempting to find the directory with the matching ID.
*/
func (d *Directory) Walk(dirID string) *Directory {
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", d.Name, d.ID)
		return nil
	}
	return walk(d, dirID)
}

func walk(dir *Directory, dirID string) *Directory {
	if len(dir.Dirs) == 0 {
		return nil
	}
	if dir.ID == dirID {
		return dir
	}
	for _, subDirs := range dir.Dirs {
		if sd := walk(subDirs, dirID); sd != nil {
			return sd
		}
	}
	return nil
}

/*
WalkS recursively traverses each subdirectory starting from
the given directory and returns a *SyncIndex pointer containing
the last sync times for each file in each directory and subdirectories.
*/
func (d *Directory) WalkS() *SyncIndex {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no files", d.Name, d.ID)
	}
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", d.Name, d.ID)
		return nil // nothing to search
	}
	return walkS(d, NewSyncIndex())
}

func walkS(dir *Directory, idx *SyncIndex) *SyncIndex {
	// check files
	if len(dir.Files) > 0 {
		for _, file := range dir.Files {
			idx.LastSync[file.ServerPath] = file.LastSync
		}
	} else {
		log.Printf("[DEBUG] dir %s (%s) has no files", dir.Name, dir.ID)
	}
	// check sub directories
	if len(dir.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", dir.Name, dir.ID)
		return idx
	}
	for _, subDirs := range dir.Dirs {
		if sIndx := walkS(subDirs, idx); sIndx != nil {
			return sIndx
		}
	}
	return nil
}

// TODO: test!

// Walkf() searches each subdirectory recursively and performes
// a supplied function on each file in the directory, returning
// an error if the function fails
func (d *Directory) WalkF(op func(file *File) error) {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no files", d.Name, d.ID)
	}
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", d.Name, d.ID)
		return
	}
	walkf(d, op)
}

func walkf(dir *Directory, op func(file *File) error) {
	// run operation on each file, if possible
	if len(dir.Files) > 0 {
		for _, file := range dir.Files {
			if err := op(file); err != nil {
				log.Printf("[DEBUG] unable to run operation on %s \n%v\n continuing...", dir.Name, err)
			}
		}
	} else {
		log.Printf("[DEBUG] dir %s (%s) has no files", dir.Name, dir.ID)
	}
	// search subdirectories
	if len(dir.Dirs) == 0 {
		return
	}
	for _, subDirs := range dir.Dirs {
		walkf(subDirs, op)
	}
}
