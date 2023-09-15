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

// create a new root directory object. does not create physical directory.
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
		LastSync:  time.Now().UTC(),
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
		LastSync:  time.Now().UTC(),
		Dirs:      make(map[string]*Directory, 0),
		Files:     make(map[string]*File, 0),
		Root:      false,
		Path:      path,
	}
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
	if !d.Protected {
		d.clear()
	} else {
		if password == d.Key {
			d.clear()
		} else {
			log.Print("[DEBUG] wrong password. contents not deleted")
		}
	}
}

// clean all files and subdirectories from the top-level directory
func clean(dirPath string) error {
	d, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("[ERROR] unable to read directory: %v", err)
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dirPath, name)); err != nil {
			return fmt.Errorf("[ERROR] unable to remove file: %v", err)
		}
	}

	return nil
}

// calls d.clean() which removes all files and subdirectories
// from a drive starting at the given path
func (d *Directory) Clean(dirPath string) error {
	if !d.Protected {
		if err := clean(dirPath); err != nil {
			return err
		}
		return nil
	} else {
		log.Printf("[DEBUG] drive is protected.")
	}
	return nil
}

func (d *Directory) HasFile(fileID string) bool {
	if _, ok := d.Files[fileID]; ok {
		return true
	}
	return false
}

func (d *Directory) HasDir(dirID string) bool {
	if _, ok := d.Dirs[dirID]; ok {
		return true
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

// updates internal file map and file's sync time
func (d *Directory) addFile(file *File) {
	d.Files[file.ID] = file
	d.Files[file.ID].LastSync = time.Now().UTC()

	log.Printf("[DEBUG] file %s (%s) added", file.Name, file.ID)
}

func (d *Directory) AddFile(file *File) {
	if !d.Protected {
		if !d.HasFile(file.ID) {
			d.addFile(file)
		} else {
			log.Printf("[DEBUG] file %s (%s) already present in directory", file.Name, file.ID)
		}
	} else {
		log.Printf("[DEBUG] directory %s (%s) locked", d.Name, d.ID)
	}
}

func (d *Directory) AddFiles(files []*File) {
	if len(files) == 0 {
		log.Printf("[DEBUG] no files recieved")
		return
	}
	if !d.Protected {
		for _, f := range files {
			if !d.HasFile(f.ID) {
				d.addFile(f)
			} else {
				log.Printf("[DEBUG] file (%v) already exists)", f.ID)
			}
		}
	} else {
		log.Printf("[DEBUG] directory %s (%s) locked", d.Name, d.ID)
	}
}

// removes internal file object from file map
func (d *Directory) removeFile(fileID string) error {
	if file, ok := d.Files[fileID]; ok {
		if err := os.Remove(file.Path); err != nil {
			log.Fatalf("[ERROR] unable to remove file: %v", err)
		}
		delete(d.Files, file.ID)
		d.LastSync = time.Now().UTC()

		log.Printf("[DEBUG] file %s removed", file.ID)
	} else {
		return fmt.Errorf("[ERROR] file %s not found", file.ID)
	}
	return nil
}

// Removes actual file plus internal File object
func (d *Directory) RemoveFile(fileID string) error {
	if !d.Protected {
		if err := d.removeFile(fileID); err != nil {
			return fmt.Errorf("[ERROR] unable to remove file: %s", err)
		}
	} else {
		log.Printf("[DEBUG] directory protected. unlock before removing files")
	}
	return nil
}

// return a copy of the files map
func (d *Directory) GetFiles() map[string]*File {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir (%s) has no files", d.ID)
	}
	return d.Files
}

// -------- sub directory methods

// creates the directory and updates internal data structures
func (d *Directory) addSubDir(dir *Directory) error {
	if err := os.MkdirAll(dir.Path, PERMS); err != nil {
		return fmt.Errorf("[ERROR] could not create directory: %v", err)
	}
	dir.Parent = d
	d.Dirs[dir.ID] = dir
	d.Dirs[dir.ID].LastSync = time.Now().UTC()

	log.Printf("[DEBUG] dir %s (%s) added", dir.Name, dir.ID)
	return nil
}

// add a single sub directory to the current directory
//
// creates a new directory using dir's internal file path and
// initializes a directory structure
func (d *Directory) AddSubDir(dir *Directory) error {
	if !d.Protected {
		if err := d.addSubDir(dir); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] dir %s is protected", d.Name)
	}
	return nil
}

// add a slice of directory objects to the current directory
func (d *Directory) AddSubDirs(dirs []*Directory) error {
	if len(dirs) == 0 {
		log.Printf("[DEBUG] dir list is empty")
		return nil
	}
	if !d.Protected {
		for _, dir := range dirs {
			if err := d.addSubDir(dir); err != nil {
				return err
			}
		}
	} else {
		log.Printf("[DEBUG] directory (id=%s) is protected", d.ID)
	}
	return nil
}

func (d *Directory) removeDir(dirID string) error {
	if dir, ok := d.Dirs[dirID]; ok {
		if err := os.Remove(dir.Path); err != nil {
			return fmt.Errorf("[ERROR] unable to remove directory %s: %v", dirID, err)
		}
	} else {
		log.Printf("[DEBUG] directory (id=%s) is not found", dirID)
	}
	return nil
}

// removes a subdirecty and *all of its child directories*
// use with caution!
func (d *Directory) RemoveSubDir(dirID string) error {
	if !d.Protected {
		if err := d.removeDir(dirID); err != nil {
			return err
		}
		// remove from subdir map & update sync time
		delete(d.Dirs, dirID)
		d.LastSync = time.Now().UTC()

		log.Printf("[DEBUG] directory %s deleted", dirID)
	} else {
		log.Printf("[DEBUG] directory %s is protected", dirID)
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
		d.Clear(d.Key)
		log.Printf("[DEBUG] dir(%s) all sub directories deleted", d.ID)
	} else {
		log.Printf("[DEBUG] dir(%s) is protected. no sub directories deleted", d.ID)
	}
	return nil
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
//
// TODO: implement our own version of Walk for this function
func (d *Directory) DirSize() (float64, error) {
	var size float64

	err := filepath.Walk(d.Path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// TODO: investigate how the conversion of
			// int64 to float64 can effect results.
			size += float64(info.Size())
		}
		return nil
	})

	return size, err
}

/*
Walk() recursively traverses sub directories starting at a given directory (or root),
attempting to find the desired file with the given file ID.
*/
func (d *Directory) WalkF(fileID string) *File {
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories. nothing to search", d.Name, d.ID)
		return nil
	}
	return walkF(d, fileID)
}

func walkF(dir *Directory, fileID string) *File {
	if len(dir.Dirs) == 0 {
		return nil
	}
	if f, found := dir.Files[fileID]; found {
		return f
	}
	for _, subDirs := range dir.Dirs {
		if sd := walkF(subDirs, fileID); sd != nil {
			return sd
		}
	}
	return nil
}

/*
Walk() recursively traverses sub directories starting at a given directory (or root),
attempting to find the desired sub directory with the given directory ID.
*/
func (d *Directory) WalkD(dirID string) *Directory {
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories. nothing to search", d.Name, d.ID)
		return nil
	}
	return walkD(d, dirID)
}

func walkD(dir *Directory, dirID string) *Directory {
	if len(dir.Dirs) == 0 {
		return nil
	}
	if dir.ID == dirID {
		return dir
	}
	for _, subDirs := range dir.Dirs {
		if sd := walkD(subDirs, dirID); sd != nil {
			return sd
		}
	}
	return nil
}

/*
WalkS() recursively traverses each subdirectory starting from
the given directory and returns a *SyncIndex pointer containing
the last sync times for each file in each directory and subdirectories.
*/
func (d *Directory) WalkS() *SyncIndex {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no files", d.Name, d.ID)
	}
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories. nothing to search.", d.Name, d.ID)
		return nil // nothing to search
	}
	// TODO: get user and userID to pass to NewSyncIndex
	return walkS(d, NewSyncIndex(d.Owner))
}

func walkS(dir *Directory, idx *SyncIndex) *SyncIndex {
	// check files
	if len(dir.Files) > 0 {
		for _, file := range dir.Files {
			idx.LastSync[file.ID] = file.LastSync
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

// d.WalkU() populates the ToUpdate map of a given SyncIndex
func (d *Directory) WalkU(idx *SyncIndex) *SyncIndex {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no files", d.Name, d.ID)
	}
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories. nothing to search.", d.Name, d.ID)
		return idx
	}
	return walkU(d, idx)
}

// walkU recursively walks the directory tree and checks the last sync time
// of each file in each subdirectory, populating the ToUpdate map of a given SyncIndex
// as needed.
func walkU(dir *Directory, idx *SyncIndex) *SyncIndex {
	// check files
	if len(dir.Files) > 0 {
		for _, file := range dir.Files {
			if _, exists := idx.LastSync[file.ID]; exists {
				// check if the time difference between most recent sync
				// and last sync is greater than zero.
				if file.LastSync.Sub(idx.LastSync[file.ID]) > 0 {
					idx.ToUpdate[file.ID] = file
				}
			} else {
				continue // this wasn't found previously, ignore
			}
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
		if uIndx := walkU(subDirs, idx); uIndx != nil {
			return uIndx
		}
	}
	return nil
}

// WalkO() searches each subdirectory recursively and performes
// a supplied function on each file in the directory, returning
// an error if the function fails
//
// functions should have the following signature: func(file *File) error
func (d *Directory) WalkO(op func(file *File) error) error {
	if len(d.Files) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no files", d.Name, d.ID)
	}
	if len(d.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", d.Name, d.ID)
		return nil
	}
	return walkO(d, op)
}

func walkO(dir *Directory, op func(f *File) error) error {
	// run operation on each file, if possible
	if len(dir.Files) > 0 {
		for _, file := range dir.Files {
			if err := op(file); err != nil {
				// we don't exit right away because this exception may only apply
				// to a single file.
				log.Printf("[DEBUG] unable to run operation on %s \n%v\n continuing...", dir.Name, err)
				continue
			}
		}
	} else {
		log.Printf("[DEBUG] dir %s (%s) has no files", dir.Name, dir.ID)
	}
	// search subdirectories
	if len(dir.Dirs) == 0 {
		log.Printf("[DEBUG] dir %s (%s) has no sub directories", dir.Name, dir.ID)
		return nil
	}
	for _, subDirs := range dir.Dirs {
		if err := walkO(subDirs, op); err != nil {
			return fmt.Errorf("failed to walk: %v", err)
		}
	}
	return nil
}
