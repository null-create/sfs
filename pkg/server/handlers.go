package server

import (
	"fmt"
	"net/http"
	"os"
)

/*
Handlers for directly working with sfs service instance.

These will likely be called by middleware, which will themselves
be passed to the router when it is instantiated.

We want to add some middleware above these calls to handle user au
and other such business to validate requests to the server.
*/

func (s *Server) GetFileInfo(fileID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			f, err := findFile(fileID, s.Svc.DbDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			data, err := f.ToJSON()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(data)
		}
	}
}

// retrieve a file from server.
//
// fileID should be parsed from the URL as a parameter
func (s *Server) GetFile(fileID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			f, err := findFile(fileID, s.Svc.DbDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// Set the response header for the download
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
			w.Header().Set("Content-Type", "application/octet-stream")

			http.ServeFile(w, r, f.ServerPath)
		}
	}
}

// send a file to server
//
// fileID should be parsed from the URL as a parameter
func (s *Server) PutFile(fileID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// get file metadata from server, if it exists
			f, err := findFile(fileID, s.Svc.DbDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// FormFile returns the first file for the given key `myFile`
			// it also returns the FileHeader so we can get the Filename,
			// the Header and the size of the file
			formFile, header, err := r.FormFile("myFile")
			if err != nil {
				ServerErr(w, fmt.Sprintf("failed to retrive form file data: %v", err))
				return
			}
			defer formFile.Close()

			// open (or create) the servers file for this user
			serverFile, err := os.Create(f.ServerPath)
			if err != nil {
				ServerErr(w, fmt.Sprintf("failed to create or truncate file: %v", err))
				return
			}
			defer serverFile.Close()

			// write file contents to server's pysical file
			data := make([]byte, header.Size)
			_, err = formFile.Read(data)
			if err != nil {
				ServerErr(w, fmt.Sprintf("failed to read file contents: %v", err))
				return
			}
			_, err = serverFile.Write(data)
			if err != nil {
				ServerErr(w, fmt.Sprintf("failed to save file to server: %v", err))
				return
			}
		}
	}
}

// attempts to read data from the user database.
//
// if found, it will attempt to prepare it as json data and return it
func (s *Service) getUser(userID string) ([]byte, error) {
	u, err := findUser(userID, s.DbDir)
	if err != nil {
		return nil, err
	}
	jsonData, err := u.ToJSON()
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *Server) GetUser(userID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// attempt to get user from user map.
		// if unsuccessful, attempt to read from user database
		if u, exist := s.Svc.Users[userID]; !exist {
			userData, err := s.Svc.getUser(userID)
			if err != nil {
				ServerErr(w, "failed to retrieve user data")
			}
			w.Write(userData)
		} else {
			jsonData, err := u.ToJSON()
			if err != nil {
				ServerErr(w, "failed to retrieve user data")
			}
			w.Write(jsonData)
		}
	}
}
