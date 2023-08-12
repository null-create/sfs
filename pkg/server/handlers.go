package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func (s *Server) DownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Open the file to be downloaded
		filePath := "example.txt" // TODO: Replace with the path to the file you want to download to
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Set the response header for the download
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name()))
		w.Header().Set("Content-Type", "application/octet-stream")

		// Copy the file's content to the response writer
		io.Copy(w, file)
	}
}

func (s *Server) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Get the uploaded file from the request
			file, handler, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "Error uploading file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Create a new file on the server to save the uploaded content
			uploadedFilePath := handler.Filename
			newFile, err := os.Create(uploadedFilePath)
			if err != nil {
				http.Error(w, "Error creating file", http.StatusInternalServerError)
				return
			}
			defer newFile.Close()

			// Copy the uploaded content to the new file
			io.Copy(newFile, file)

			fmt.Fprintf(w, "File uploaded successfully")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) GetFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) PutFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) GetFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) GetDir() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) GetDirs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// func (s *Server) handleTemplate(files string ...) http.HandlerFunc {
// 	var (
//        // a one time use initialization controller. used for optimization purposes. see: https://pkg.go.dev/sync#example-Once
// 			init    sync.Once
// 			tpl     *template.Template
// 			tplerr  error
// 	)
// 	return func(w http.ResponseWriter, r *http.Request) {
// 			init.Do(func(){
// 					tpl, tplerr = template.ParseFiles(files...)
// 			})
// 			if tplerr != nil {
// 					http.Error(w, tplerr.Error(), http.StatusInternalServerError)
// 					return
// 			}
// 			// use tpl
// 	}
// }

// middleware stuff
// func (s *Server) adminOnly(h http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if !currentUser(r).IsAdmin {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		h(w, r)
// 	}
// }
