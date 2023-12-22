package service

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// enusre -rw-r----- permissions
const PERMS = 0640 // go's default is 0666

type ChecksumMismatch error

type File struct {
	m sync.Mutex

	// metadata
	ID      string  `json:"id"`       // file id
	Name    string  `json:"name"`     // file name
	NMap    NameMap `json:"name_map"` // file name map
	OwnerID string  `json:"owner"`    // file owner ID
	DirID   string  `json:"dir_id"`   // id of the directory this file belongs to

	// security stuff
	Protected bool   `json:"protected"`
	Key       string `json:"key"`

	// synchronization and file integrity fields
	LastSync   time.Time `json:"last_sync"`
	Path       string    `json:"path"` // temp. will be replaced by server/client path at some point
	ServerPath string    `json:"server_path"`
	ClientPath string    `json:"client_path"`
	Endpoint   string    `json:"endpoint"` // unique API endpoint

	CheckSum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`

	// file content/bytes
	Content []byte
}

// creates a new file struct instance.
// file contents are not loaded into memory.
func NewFile(fileName string, ownerID string, path string) *File {
	cs, err := CalculateChecksum(path, "sha256")
	if err != nil {
		log.Printf("[WARNING] error calculating checksum: %v", err)
	}

	uuid := NewUUID()
	cfg := NewSvcCfg()

	return &File{
		Name:       fileName,
		ID:         uuid,
		NMap:       newNameMap(fileName, uuid),
		OwnerID:    ownerID,
		Protected:  false,
		Key:        "default",
		LastSync:   time.Now().UTC(),
		Path:       path,
		ServerPath: path,
		ClientPath: path,
		Endpoint:   fmt.Sprint(Endpoint, ":", cfg.Port, "/v1/files/", uuid),
		CheckSum:   cs,
		Algorithm:  "sha256",
		Content:    make([]byte, 0),
	}
}

func UnmarshalFileStr(fileInfo string) (*File, error) {
	file := new(File)
	if err := json.Unmarshal([]byte(fileInfo), &file); err != nil {
		return nil, err
	}
	return file, nil
}

// returns file size in bytes
//
// uses os.Stat() - "length in bytes for regular files; system-dependent for others"
func (f *File) Size() int64 {
	info, err := os.Stat(f.Path)
	if err != nil {
		log.Fatalf("unable to determine file size: %v", err)
	}
	return info.Size()
}

// convert file object to json-formatted byte slice
func (f *File) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
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

func (f *File) ChangePassword(password string, newPassword string) {
	if password == f.Key {
		f.Key = newPassword
		log.Printf("[INFO] %s (id=%s) password updated!", f.Name, f.ID)
	} else {
		log.Print("[INFO] wrong password")
	}
}

// ----------- I/O

func (f *File) Load() {
	if f.Path == "" {
		log.Fatalf("no path specified")
	}
	if !f.Protected {
		f.m.Lock()
		defer f.m.Unlock()

		file, err := os.Open(f.Path)
		if err != nil {
			log.Fatalf("unable to open file %s: %v", f.Path, err)
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

		file, err := os.Create(f.Path)
		if err != nil {
			return fmt.Errorf("unable to create file %s: %v", f.Name, err)
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("unable to write file %s: %v", f.Path, err)
		}
		if err := f.UpdateChecksum(); err != nil {
			return fmt.Errorf("failed to update checksum: %v", err)
		}
		f.LastSync = time.Now().UTC()
	} else {
		log.Printf("[INFO] %s is protected", f.Name)
	}
	return nil
}

// clears *f.Content* not the actual external file contents!
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
func (f *File) Copy(dst string) error {
	s, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
	return nil
}

// ----------- File integrity

func CalculateChecksum(filePath string, hashType string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var h hash.Hash
	switch hashType {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	default:
		return "", fmt.Errorf("unsupported hash type: %s", hashType)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	checksum := string(h.Sum(nil))
	return checksum, nil
}

func (f *File) ValidateChecksum() error {
	cs, err := CalculateChecksum(f.Path, f.Algorithm)
	if err != nil {
		return fmt.Errorf("unable to calculate checksum: %v", err)
	}
	if cs != f.CheckSum {
		return fmt.Errorf("checksum mismatch! orig: %s, new: %s", cs, f.CheckSum)
	}
	return nil
}

func (f *File) UpdateChecksum() error {
	newCs, err := CalculateChecksum(f.Path, f.Algorithm)
	if err != nil {
		return fmt.Errorf("CalculateChecksum failed: %v", err)
	}
	f.CheckSum = newCs
	f.LastSync = time.Now().UTC()
	return nil
}
