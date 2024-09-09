package service

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/logger"
)

type File struct {
	m sync.Mutex

	// metadata
	ID           string      `json:"id"`            // file id
	Name         string      `json:"name"`          // file name
	NMap         NameMap     `json:"name_map"`      // file name map
	OwnerID      string      `json:"owner"`         // file owner ID
	DirID        string      `json:"dir_id"`        // id of the directory this file belongs to
	DriveID      string      `json:"drive_id"`      // id of the drive this file belongs to
	Mode         fs.FileMode `json:"mode"`          // file permissions
	Size         int64       `json:"size"`          // file size in bytes
	ServerBackup bool        `json:"server_backup"` // flag for whether this is the server-side version of the file
	LocalBackup  bool        `json:"local_backup"`  // flag for whether this is a local backup version of the file
	Registered   bool        `json:"registered"`    // flag for whether the file has been registered with the server

	// security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"-"`

	// synchronization and file integrity fields
	LastSync   time.Time `json:"last_sync"`   // last sync time for this file
	Path       string    `json:"path"`        // temp. will be replaced by server/client path at some point
	ServerPath string    `json:"server_path"` // path to file on the server
	ClientPath string    `json:"client_path"` // path to file on the client
	BackupPath string    `json:"backup_path"` // path to local backup version of this file
	Endpoint   string    `json:"endpoint"`    // unique server API endpoint
	CheckSum   string    `json:"checksum"`    // file checksum
	Algorithm  string    `json:"algorithm"`   // checksum algorithm

	// file content
	Content []byte
}

// creates a new file struct instance.
// file contents are not loaded into memory.
func NewFile(fileName string, driveID string, ownerID string, filePath string) *File {
	// for logging any errors during new file object creation
	var nfLog = logger.NewLogger("FILE_INIT", "None")

	// service configs
	cfg := NewSvcCfg()

	// get baseline information about the file
	item, err := os.Stat(filePath)
	if err != nil {
		nfLog.Error(fmt.Sprintf("failed to get file stats: %v", err))
		log.Fatal(fmt.Errorf("failed to get file stats: %v", err))
	}
	if item.IsDir() {
		nfLog.Error(fmt.Sprintf("item is a directory: %v", filePath))
		log.Fatal(fmt.Errorf("item is a directory: %v", filePath))
	}
	cs, err := CalculateChecksum(filePath)
	if err != nil {
		nfLog.Warn("error calculating checksum: " + err.Error())
	}

	// assign new id
	uuid := auth.NewUUID()
	return &File{
		Name:         fileName,
		ID:           uuid,
		NMap:         newNameMap(fileName, uuid),
		OwnerID:      ownerID,
		DriveID:      driveID,
		Mode:         item.Mode(),
		Size:         item.Size(),
		ServerBackup: false,
		Protected:    false,
		Key:          auth.GenSecret(64),
		LastSync:     time.Now().UTC(),
		Path:         filePath,
		ServerPath:   filePath,
		ClientPath:   filePath,
		BackupPath:   filePath,
		Registered:   false,
		Endpoint:     Endpoint + ":" + cfg.Port + "/v1/files/" + uuid,
		CheckSum:     cs,
		Algorithm:    "sha256",
		Content:      make([]byte, 0),
	}
}

// has this file been backed up ?
func (f *File) IsServerBackUp() bool  { return f.ServerBackup }
func (f *File) IsLocalBackup() bool   { return f.LocalBackup }
func (f *File) GetBackupPath() string { return f.BackupPath }

func UnmarshalFileStr(fileInfo string) (*File, error) {
	file := new(File)
	if err := json.Unmarshal([]byte(fileInfo), &file); err != nil {
		return nil, err
	}
	return file, nil
}

// retrieve a string in JSON format containing the files metadata.
func (f *File) ToString() string {
	data, err := f.ToJSON()
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

// convert file object to json-formatted byte slice
func (f *File) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// returns file size in bytes
//
// uses os.Stat() - "length in bytes for regular files; system-dependent for others"
func (f *File) GetSize() int64 {
	info, err := os.Stat(f.GetPath())
	if err != nil {
		log.Fatalf("unable to determine file size: %v", err)
	}
	return info.Size()
}

// Get the size of a file as a string
func (f *File) GetSizeStr() string {
	return strconv.Itoa(int(f.GetSize()))
}

// confirms the physical file associated with this object actually exists.
func (f *File) Exists() bool {
	if _, err := os.Stat(f.GetPath()); errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		log.Fatal(err)
	}
	return true
}

// mark this object as being the server side version of the original file.
func (f *File) MarkServerBackUp() {
	if !f.IsServerBackUp() {
		f.ServerBackup = true
	}
}

// mark this file as being the local back up version of the original file
func (f *File) MarkLocalBackup() {
	if !f.IsLocalBackup() {
		f.LocalBackup = true
	}
}

// get the path for this file.
// this will return either the client or server path, depending on instance called.
// server side files will have file.Backup set to true, client side files will not.
func (f *File) GetPath() string {
	var path string
	if f.ServerBackup {
		path = f.ServerPath // server side file path
	} else {
		path = f.ClientPath // client side file path (original)
	}
	return path
}

// ----------- simple security features

func (f *File) Lock(password string) {
	if password == f.Key {
		f.Protected = true
	} else {
		log.Print("[INFO] wrong password")
	}
}

func (f *File) Unlock(password string) {
	if password == f.Key {
		f.Protected = false
	} else {
		log.Print("[INFO] wrong password")
	}
}

func (f *File) ChangePassword(oldPassword string, newPassword string) {
	if oldPassword == f.Key {
		f.Key = newPassword
		log.Printf("[INFO] %s (id=%s) password updated!", f.Name, f.ID)
	} else {
		log.Print("[INFO] wrong password")
	}
}

// ----------- I/O

// load file contents into memory
func (f *File) Load() {
	if f.GetPath() == "" {
		log.Fatalf("no path specified")
	}
	if !f.Protected {
		f.m.Lock()
		defer f.m.Unlock()

		file, err := os.Open(f.GetPath())
		if err != nil {
			log.Fatalf("unable to open file %s: %v", f.GetPath(), err)
		}
		defer file.Close()

		data, err := os.ReadFile(file.Name())
		if err != nil {
			log.Fatalf("unable to read file %s: %v", f.Name, err)
		}
		f.Content = data
	} else {
		log.Printf("[INFO] %s is protected", f.Name)
	}
}

// update (or create) a file. updates checksum and last sync time.
func (f *File) Save(data []byte) error {
	if !f.Protected {
		f.m.Lock()
		defer f.m.Unlock()

		file, err := os.Create(f.GetPath())
		if err != nil {
			return fmt.Errorf("unable to create file %s: %v", f.Name, err)
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("unable to write file %s: %v", f.GetPath(), err)
		}
		if err := f.UpdateChecksum(); err != nil {
			return fmt.Errorf("failed to update checksum: %v", err)
		}
		f.Size = f.GetSize()
		f.LastSync = time.Now().UTC()
	} else {
		log.Printf("[INFO] %s is protected", f.Name)
	}
	return nil
}

// clears the *in-memory* file contents, not the actual external file contents.
func (f *File) Clear() error {
	if !f.Protected {
		f.Content = nil
		f.Content = make([]byte, 0)
		log.Printf("[INFO] %s in-memory content cleared (external file not altered)", f.Name)
	} else {
		log.Printf("[INFO] %s is protected", f.Name)
	}
	return nil
}

// copy this file to another location
func (f *File) Copy(destPath string) error {
	if !f.Protected {
		f.m.Lock()
		defer f.m.Unlock()

		src, err := os.Open(f.GetPath())
		if err != nil {
			return err
		}
		defer src.Close()

		dest, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[WARNING] %s is protected. copy not executed.", f.Name)
	}
	return nil
}

// ----------- File integrity

func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var h = sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	checksum := base32.StdEncoding.EncodeToString(h.Sum(nil))
	return checksum, nil
}

func (f *File) ValidateChecksum() error {
	cs, err := CalculateChecksum(f.GetPath())
	if err != nil {
		return fmt.Errorf("unable to calculate checksum: %v", err)
	}
	if cs != f.CheckSum {
		return fmt.Errorf("checksum mismatch! orig: %s, new: %s", cs, f.CheckSum)
	}
	return nil
}

func (f *File) UpdateChecksum() error {
	newCs, err := CalculateChecksum(f.GetPath())
	if err != nil {
		return fmt.Errorf("CalculateChecksum failed: %v", err)
	}
	f.CheckSum = newCs
	f.LastSync = time.Now().UTC()
	return nil
}
