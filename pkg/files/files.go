package files

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// enusre -rw-r----- permissions
const PERMS = 0640 // go's default is 066

// used to store the association between a file's name and its UUID
type NameMap map[string]string

func newNameMap(fileName string, uuid string) NameMap {
	nm := make(NameMap, 1)
	nm[fileName] = uuid
	return nm
}

type File struct {
	m sync.Mutex

	Name  string
	ID    string
	NMap  NameMap
	Owner string

	Protected bool
	Key       string

	LastSync time.Time

	Path       string
	ServerPath string
	ClientPath string

	CheckSum  string
	Algorithm string

	Content []byte
}

// Content is loaded elsewhere since
func NewFile(fileName string, owner string, path string) *File {
	cs, err := CalculateChecksum(path, "sha256")
	if err != nil {
		log.Printf("[DEBUG] Error calculating checksum: %v", err)
	}

	uuid := NewUUID()

	return &File{
		Name:  fileName,
		ID:    uuid,
		NMap:  newNameMap(fileName, uuid),
		Owner: owner,

		Protected: false,
		Key:       "default",

		Path:       path,
		ServerPath: path,
		ClientPath: path,

		CheckSum:  cs,
		Algorithm: "sha256",
	}
}

func (f *File) HasContent() bool {
	return len(f.Content) == 0
}

// returns file size in kb
func (f *File) Size() int {
	info, err := os.Stat(f.Path)
	if err != nil {
		log.Fatalf("[ERROR] unable to determine file size: %v", err)
	}
	return int(info.Size()) / 1024.0
}

// --- simple security features

func (f *File) IsProtected() bool {
	return f.Protected
}

func (f *File) Lock(password string) {
	if password == f.Key {
		f.Protected = true
	} else {
		log.Print("[DEBUG] wrong password")
	}
}

func (f *File) Unlock(password string) {
	if password == f.Key {
		f.Protected = false
	} else {
		log.Print("[DEBUG] wrong password")
	}
}

func (f *File) ChangePassword(password string, newPassword string) {
	if password == f.Key {
		f.Key = newPassword
		log.Print("[DEBUG] password updated!")
	} else {
		log.Print("[DEBUG] wrong password")
	}
}

//---- CRUD-------

func (f *File) Load() {
	if f.Path == "" {
		log.Fatalf("[ERROR] no path specified")
	}

	f.m.Lock()
	defer f.m.Unlock()

	file, err := os.Open(f.Path)
	if err != nil {
		log.Fatalf("[ERROR] unable to open file %s: %v", f.Path, err)
	}
	defer file.Close()

	data, err := os.ReadFile(file.Name())
	if err != nil {
		log.Fatalf("[ERROR] unable to read file %s: %v", f.Name, err)
	}
	f.Content = data
}

func (f *File) Save(data []byte) error {
	if !f.Protected {
		f.m.Lock()
		defer f.m.Unlock()

		// If the file doesn't exist, it will be created.
		file, err := os.Create(f.Path)
		if err != nil {
			return fmt.Errorf("[ERROR] unable to create file: %v", err)
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			log.Fatalf("[ERROR] unable to write file %s: %v", f.Path, err)
		}

		f.Content = data
		f.LastSync = time.Now()

	} else {
		log.Print("[DEBUG] file is protected")
	}
	return nil
}

// clears *f.Content* not the actual external file contents
func (f *File) Clear() error {
	if !f.Protected {
		f.Content = []byte{}
	} else {
		log.Print("[DEBUG] file is protected")
	}
	return nil
}

// ------- File integrity

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

	checksum := fmt.Sprintf("%x", h.Sum(nil))
	return checksum, nil
}

func (f *File) ValidateChecksum() {
	cs, err := CalculateChecksum(f.Path, f.Algorithm)
	if err != nil {
		log.Printf("[DEBUG] unable to calculate checksum: %v", err)
		return
	}
	if cs != f.CheckSum {
		log.Printf("[WARNING] checksum mismatch! orig: %s, new: %s", cs, f.CheckSum)
	}
}
