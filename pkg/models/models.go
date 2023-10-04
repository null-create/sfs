package models

/*
This module defines REST request and response objects.

This will be used by both Client and Server when creating messages
*/

type Msgs interface {
	ToJSON() ([]byte, error)
}

type ClientMessage struct{}

type ServerMessage struct{}

type FileUploadRequest struct{}

type FileUpdateRequest struct{}

type FileDownloadResponse struct{}

type DirDownloadRequest struct{}

type DirUploadRequest struct{}

type SyncRequest struct{}
