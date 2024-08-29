package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/server"
	svc "github.com/sfs/pkg/service"
)

// API handlers for the web client UI

// enums mainly used for creating context keys
type Contexts string

const Error Contexts = "error"

var (
	homePage  = "http://" + cfgs.Addr // web ui home page
	userPage  = homePage + "/user"    // users home page
	errorPage = homePage + "/error"
)

// -------- various pages -------------------------------------

func (c *Client) error(w http.ResponseWriter, r *http.Request, msg string) {
	c.log.Error(msg)
	errCtx := context.WithValue(r.Context(), server.Error, msg)
	c.ErrorPage(w, r.WithContext(errCtx))
}

func (c *Client) HomePage(w http.ResponseWriter, r *http.Request) {
	indexData := Index{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		UserID:     c.User.ID,
		Files:      c.Drive.GetFiles(),
		Dirs:       c.Drive.GetDirs(),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) ErrorPage(w http.ResponseWriter, r *http.Request) {
	errMsg := r.Context().Value(Error)
	if errMsg == nil {
		http.Error(w, "No error parsed from request", http.StatusInternalServerError)
		return
	}
	errMsg = errMsg.(string)
	errPageData := ErrorPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		ErrMsg:     fmt.Sprintf("Something went wrong :(\n\n%s", errMsg),
	}
	err := c.Templates.ExecuteTemplate(w, "error.html", errPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) UserPage(w http.ResponseWriter, r *http.Request) {
	usrPageData := UserPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       c.User.Name,
		UserID:     c.User.ID,
		UserName:   c.User.UserName,
		Email:      c.User.Email,
		TotalFiles: len(c.Drive.GetFiles()),
		TotalDirs:  len(c.Drive.GetDirs()),
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "user.html", usrPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) RecycleBinPage(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(c.RecycleBin)
	if err != nil {
		c.error(w, r, err.Error())
	}

	recycleBinItems := newRecycleBinItems()
	for _, entry := range entries {
		if entry.IsDir() {
			recycleBinItems.Dirs = append(recycleBinItems.Dirs, entry)
		} else {
			recycleBinItems.Files = append(recycleBinItems.Files, entry)
		}
	}

	recyclePageData := RecyclePage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Files:      recycleBinItems.Files,
		Dirs:       recycleBinItems.Dirs,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "recycled.html", recyclePageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) DirPage(w http.ResponseWriter, r *http.Request) {
	dirID := r.Context().Value(server.Directory).(string)
	dir, err := c.GetDirectoryByID(dirID)
	if dir == nil {
		c.error(w, r, err.Error())
		return
	}
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// get sub directories
	subdirs, err := c.GetSubDirs(dirID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// get files
	files, err := c.GetFilesByDirID(dirID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	dirPageData := DirPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       dir.Name,
		Size:       dir.Size,
		LastSync:   dir.LastSync,
		Dirs:       subdirs,
		Files:      files,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "folder.html", dirPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) FilePage(w http.ResponseWriter, r *http.Request) {
	fileID := r.Context().Value(server.File).(string)
	file, err := c.GetFileByID(fileID)
	if file == nil {
		c.error(w, r, err.Error())
		return
	}
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	filePageData := FilePage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		Name:       file.Name,
		Size:       file.Size,
		ID:         file.ID,
		OwnerID:    file.OwnerID,
		Type:       filepath.Ext(file.Name),
		LastSync:   file.LastSync,
		Location:   file.ClientPath,
		Checksum:   file.CheckSum,
		Endpoint:   file.Endpoint,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err = c.Templates.ExecuteTemplate(w, "file.html", filePageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) RemoveFileHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	fileID := buf.String()
	file, err := c.GetFileByID(fileID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	if err := c.RemoveFile(file); err != nil {
		c.error(w, r, err.Error())
		return
	}
}

func (c *Client) handleNewUserInfo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	// Extract form data (add more fields as needed)
	// var updates = []string{
	// 	r.FormValue("name"), r.FormValue("email"),
	// 	r.FormValue("username"),
	// }

	// iterate through and if any are not empty and are different than
	// the current configurations, update in db and .env files accordingly

	// send back to the user's page once complete
	http.Redirect(w, r, userPage, http.StatusSeeOther)
}

// update client information received from the web ui
func (c *Client) EditInfo(w http.ResponseWriter, r *http.Request) {
	editPage := EditPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
	}
	err := c.Templates.ExecuteTemplate(w, "edit.html", editPage)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) AddPage(w http.ResponseWriter, r *http.Request) {
	addPageData := AddPage{
		UserPage:     userPage,
		ProfilePic:   c.Conf.ProfilePic,
		DiscoverPath: "CHANGEME",
		ServerHost:   c.Conf.ServerAddr,
		ClientHost:   c.Conf.Addr,
		Endpoint:     c.Endpoints["new-file"],
	}
	err := c.Templates.ExecuteTemplate(w, "add.html", addPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

// add a file or directory to the SFS service
func (c *Client) AddItems(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	r.Body.Close()

	var path = buf.String()
	item, err := os.Stat(path)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	if item.IsDir() {
		_, err = c.Discover(path)
		if err != nil {
			c.error(w, r, err.Error())
			return
		}
	} else {
		if err := c.AddFile(path); err != nil {
			c.error(w, r, err.Error())
			return
		}
	}

	// Redirect to the page for the newly mapped out directory
	http.Redirect(w, r, homePage, http.StatusSeeOther)
}

func (c *Client) UploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("upload handler: request: %s\n", r.Form)

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		c.error(w, r, err.Error())
		return
	}
	file, handler, err := r.FormFile("newFile")
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	defer file.Close()

	// You can save the file on the server or process it as needed
	fmt.Fprintf(w, "Uploaded File: %+v\n", handler.Filename)
	fmt.Fprintf(w, "File Size: %+v\n", handler.Size)
	fmt.Fprintf(w, "MIME Header: %+v\n", handler.Header)

	// retrieve destination directory from request
	destDirName := r.Form.Get("destDir")
	if destDirName == "" {
		c.error(w, r, "no destination directory specified")
		return
	}
	if err := c.AddFile("CHANGEME"); err != nil {
		c.error(w, r, err.Error())
		return
	}

	// redirect back to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (c *Client) UpdatePfpHandler(w http.ResponseWriter, r *http.Request) {
	// r.ParseMultipartForm(10 << 20) // limit file size to 10mb
	// file, handler, err := r.FormFile("profilePic")
	// if err != nil {
	// 	fmt.Printf("error parsing form: %v", err)
	// 	c.error(w, r, err.Error())
	// 	return
	// }
	// defer file.Close()

	// fmt.Printf("uploaded File: %v\n", handler.Filename)

	// Save the file to the client
	// destPath, err := filepath.Abs("./assets/profile-pics")
	// if err != nil {
	// 	c.error(w, r, err.Error())
	// 	return
	// }

	// fmt.Printf("saving file to: %v\n", destPath)

	// filePath := filepath.Join(destPath, handler.Filename)
	// dst, err := os.Create(filePath)
	// if err != nil {
	// 	fmt.Printf("error: %v", err)
	// 	c.error(w, r, err.Error())
	// 	return
	// }
	// defer dst.Close()

	// fmt.Printf("copying...\n")

	// // Copy the uploaded file to the destination and update client config accordingly
	// if _, err := io.Copy(dst, file); err != nil {
	// 	c.error(w, r, err.Error())
	// 	return
	// }
	err := r.ParseMultipartForm(10 << 20) // Limit file size to 10MB
	if err != nil {
		http.Error(w, "Unable to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, header, err := r.FormFile("profile-pic")
	if err != nil {
		http.Error(w, "Unable to retrieve file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a destination file on the server
	fileName := filepath.Base(header.Filename)
	dst, err := os.Create(fmt.Sprintf("./assets/profile-pics/%s", fileName))
	if err != nil {
		http.Error(w, "Unable to create the file on server: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file's contents to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Unable to save the file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send back a JSON response with the new profile picture URL
	newProfilePicURL := fmt.Sprintf("/assets/profile-pics/%s", fileName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintf(w, `{"newProfilePicURL": "%s"}`, newProfilePicURL)
	if err != nil {
		http.Error(w, "Failed to format response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.UpdateConfigSetting("CLIENT_PROFILE_PIC", fileName); err != nil {
		c.error(w, r, err.Error())
		return
	}

	fmt.Printf("redirecting...\n")

	// Redirect or respond with a success message
	http.Redirect(w, r, "/user", http.StatusSeeOther)
}

func (c *Client) ClearPfpHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.UpdateConfigSetting("CLIENT_PROFILE_PIC", "default_profile_pic.jpg"); err != nil {
		c.error(w, r, err.Error())
		return
	}
	// Redirect or respond with a success message
	http.Redirect(w, r, "/user", http.StatusNoContent)
}

// render upload page template
func (c *Client) UploadPage(w http.ResponseWriter, r *http.Request) {
	usersDirs, err := c.Db.GetUsersDirectories(c.UserID)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	uploadPageData := UploadPage{
		UserPage:   userPage,
		Dirs:       usersDirs,
		ProfilePic: c.Conf.ProfilePic,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
		Endpoint:   c.Endpoints["new file"],
	}
	err = c.Templates.ExecuteTemplate(w, "upload.html", uploadPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) SettingsPage(w http.ResponseWriter, r *http.Request) {
	settingsPageData := SettingsPage{
		UserPage:   userPage,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
		// alterable settings
		UserName:   c.Conf.User,
		UserAlias:  c.Conf.UserAlias,
		UserEmail:  c.User.Email,
		LocalSync:  c.Conf.LocalBackup,
		BackupDir:  c.Conf.BackupDir,
		ClientPort: c.Conf.Port,
	}
	err := c.Templates.ExecuteTemplate(w, "settings.html", settingsPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	// get updated settings and modify accordingly

	http.Redirect(w, r, userPage, http.StatusSeeOther)
}

// search for items
func (c *Client) SearchPage(w http.ResponseWriter, r *http.Request) {
	var resultsFile = "search-results.json"

	if r.Method == http.MethodPost {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r.Body)
		if err != nil {
			c.error(w, r, err.Error())
			return
		}
		r.Body.Close()

		searchItem := buf.String()
		files, dirs, err := c.SearchForItems(searchItem)
		if err != nil {
			c.error(w, r, err.Error())
			return
		}
		results := SearchResults{
			Files: files,
			Dirs:  dirs,
		}
		// save to temp json file so the search page
		// can display results when called with a GET request
		data, err := json.Marshal(results)
		if err != nil {
			c.error(w, r, err.Error())
			return
		}
		if err := os.WriteFile(resultsFile, data, 0644); err != nil {
			c.error(w, r, err.Error())
			return
		}

	} else if r.Method == http.MethodGet {
		var results = NewSearchResults()
		if FileExists(resultsFile) {
			data, err := os.ReadFile(resultsFile)
			if err != nil {
				c.error(w, r, err.Error())
				return
			}
			if err := json.Unmarshal(data, &results); err != nil {
				c.error(w, r, err.Error())
				return
			}
			_ = os.Remove(resultsFile)
		}
		searchPageData := SearchPage{
			UserID:     c.UserID,
			UserPage:   userPage,
			ServerHost: c.Conf.ServerAddr,
			ClientHost: c.Conf.Addr,
			Files:      results.Files,
			Dirs:       results.Dirs,
		}
		err := c.Templates.ExecuteTemplate(w, "search.html", searchPageData)
		if err != nil {
			c.error(w, r, err.Error())
		}
	}
}

// -------- files -----------------------------------------

// gets a file from the ID provided by the request. returns nil if not found.
func (c *Client) getFileFromRequest(r *http.Request) (*svc.File, error) {
	fileID := r.Context().Value(server.File).(string)
	if fileID == "" {
		return nil, fmt.Errorf("no file ID found in request")
	}
	file, err := c.Db.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file (id=%s) not found", fileID)
	}
	return file, nil
}

// add a new file to the SFS serverice using its absolute path
func (c *Client) NewFile(w http.ResponseWriter, r *http.Request) {
	newFilePath := r.Context().Value(server.File).(string)
	if err := c.AddFile(newFilePath); err != nil {
		c.error(w, r, err.Error())
		return
	}
}

// retrieve a file from the local machine
func (c *Client) ServeFile(w http.ResponseWriter, r *http.Request) {
	file, err := c.getFileFromRequest(r)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.ClientPath)
	c.log.Info(fmt.Sprintf("served file %s: %s", file.Name, file.ClientPath))
}
