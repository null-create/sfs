package files

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"time"
)

// enusre -rw-r----- permissions
const PERMS = 0640 // go's default is 066

// used to store the association between files naem and UUID
type NameMap map[string]string

func newNameMap(fileName string, uuid string) NameMap {
	nm := make(NameMap, 1)
	nm[fileName] = uuid
	return nm
}

type File struct {
	Name  string
	ID    string
	UUID  string
	NMap  NameMap
	Owner string

	IsProtected bool
	Key         string

	LastSync time.Time

	Path       string
	ServerPath string
	ClientPath string
	CheckSum   string
	Content    []byte
}

// Content is loaded elsewhere since
func NewFile(fileName string, owner string, path string) *File {
	uuid, err := calculateChecksum(path, "sha256")
	if err != nil {
		log.Printf("[DEBUG] Error calculating checksum: %v", err)
	}
	return &File{
		Name:       fileName,
		UUID:       uuid,
		NMap:       newNameMap(fileName, uuid),
		Owner:      owner,
		Path:       path,
		ServerPath: path,
		ClientPath: path,
	}
}

func (f *File) Load() {
	if f.Path == "" {

	}

}

func (f *File) Save(data []byte) error {
	return nil
}

func (f *File) Clear() error {
	return nil
}

func calculateChecksum(filePath string, hashType string) (string, error) {
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
		h = md5.New()
	case "sha256":
		h = md5.New()
	default:
		return "", fmt.Errorf("unsupported hash type: %s", hashType)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))
	return checksum, nil
}
