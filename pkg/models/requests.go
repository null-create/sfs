package models

import (
	"time"

	svc "github.com/sfs/pkg/service"
)

/*
This module defines REST request objects.

This will be used by both Client and Server when creating request messages
*/

type InfoReq struct {
	RequestTime time.Time `json:"request_time"`

	// requestor info
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// --- info

type FileInfoReq struct {
	Req    InfoReq `json:"request"`
	FileID string  `json:"file_id"`
}

type DirInfoReq struct {
	Req   InfoReq `json:"request"`
	DirID string  `json:"dir_id"`
}

type DriveInfoReq struct {
	Req     InfoReq `json:"request"`
	DriveID string  `json:"drive_id"`
}

type UserInfoReq struct {
	Req    InfoReq `json:"request"`
	UserID string  `json:"user_id"`
}

// --- upload/download/delete/create requests

type OpReq struct {
	RequestTime time.Time `json:"request_time"`

	// requester
	User   string `json:"user"`
	UserID string `json:"user_id"`

	// file/directory/drive paths
	ClientPath string `json:"client_path"`
	ServerPath string `json:"server_path"`
}

type FileReq struct {
	Req      OpReq  `json:"request"`
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type DirReq struct {
	Req     OpReq  `json:"request"`
	DirID   string `json:"dir_id"`
	DirName string `json:"dir_name"`
}

type DriveReq struct {
	Req     OpReq  `json:"request"`
	DriveID string `json:"drive_id"`
	DrvName string `json:"drv_name"`
}

type UserReq struct {
	Req    OpReq  `json:"request"`
	UserID string `json:"user_id"`
}

// used specfically for creating new users and drives
type NewUserReq struct {
	RequestTime time.Time `json:"request_time"`

	Name     string `json:"name"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

// --- sync operations

type SyncReq struct {
	RequestTime time.Time `json:"request_time"`

	// requester
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// request to generate a new sync index on either the client
// or server
type StartIdx struct {
	Req SyncReq `json:"request"`

	// who should be building the new sync index (client or server)
	Target string `json:"target"`
}

// sync the server to the client's files
type SyncUpReq struct {
	Req       SyncReq       `json:"request"`
	ClientIdx svc.SyncIndex `json:"client_idx"`
}

// sync the client to the server's files
type SyncDownReq struct {
	Req       SyncReq       `json:"request"`
	ServerIdx svc.SyncIndex `json:"server_idx"`
}
