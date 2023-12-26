package transfer

import (
	"encoding/json"
	"log"
)

// TODO: define an interface for these contexts since they all are using the
// same util function names

// struct to use for populating context.Context
type FileContext struct {
	Name     string `json:"name"`
	OwnerID  string `json:"owner_id"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

func NewFileContext(name string, ownerID string, path string, cs string) FileContext {
	return FileContext{
		Name:     name,
		OwnerID:  ownerID,
		Path:     path,
		Checksum: cs,
	}
}

func (fctx FileContext) IsEmpty() bool {
	return fctx.Name == "" && fctx.OwnerID == "" && fctx.Path == "" && fctx.Checksum == ""
}

func (fctx FileContext) ToJSON() []byte {
	data, err := json.MarshalIndent(fctx, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// struct for containing user context within requests
type UserContext struct {
	Name    string `json:"name"`
	Alias   string `json:"alias"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

func NewUserContext(name string, alias string, email string, isAdmin bool) UserContext {
	return UserContext{
		Name:    name,
		Alias:   alias,
		Email:   email,
		IsAdmin: isAdmin,
	}
}

func (uctx UserContext) IsEmpty() bool {
	return uctx.Name == "" && uctx.Alias == "" && uctx.Email == ""
}

func (uctx UserContext) ToJSON() []byte {
	data, err := json.MarshalIndent(uctx, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	return data
}

type DirectoryContext struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
	DriveID string `json:"drive_id"`
	Path    string `json:"path"`
}

func NewDirectoryContext(id string, name string, ownerID string, driveID string, path string) DirectoryContext {
	return DirectoryContext{
		ID:      id,
		Name:    name,
		OwnerID: ownerID,
		DriveID: driveID,
		Path:    path,
	}
}

func (dc DirectoryContext) IsEmpty() bool {
	return dc.ID == "" && dc.Name == "" && dc.OwnerID == "" && dc.DriveID == "" && dc.Path == ""
}

func (dc DirectoryContext) ToJSON() []byte {
	data, err := json.MarshalIndent(dc, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	return data
}
