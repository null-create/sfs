package client

import (
	"context"
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

func (c *Client) hasPfPics() bool {
	path, err := filepath.Abs("../assets/pfp")
	if err != nil {
		c.log.Error(err.Error())
		return false
	}
	entires, err := os.ReadDir(path)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}
	return len(entires) != 0
}

func (c *Client) error(w http.ResponseWriter, r *http.Request, msg string) {
	errCtx := context.WithValue(r.Context(), server.Error, msg)
	http.Redirect(w, r.WithContext(errCtx), errorPage, http.StatusInternalServerError)
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
		UserPage:       userPage,
		ProfilePic:     c.Conf.ProfilePic,
		Name:           c.User.Name,
		UserID:         c.User.ID,
		UserName:       c.User.UserName,
		Email:          c.User.Email,
		TotalFiles:     len(c.Drive.GetFiles()),
		TotalDirs:      len(c.Drive.GetDirs()),
		ProfilePicPath: "",
		ServerHost:     c.Conf.ServerAddr,
		ClientHost:     c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "user.html", usrPageData)
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
	}
	err := c.Templates.ExecuteTemplate(w, "add.html", addPageData)
	if err != nil {
		c.error(w, r, err.Error())
	}
}

func (c *Client) AddItems(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	_, handler, err := r.FormFile("folder-path")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		c.error(w, r, err.Error())
		return
	}

	fmt.Printf("path received: %s\n", handler.Filename)

	newDir, err := c.Discover(handler.Filename)
	if err != nil {
		c.error(w, r, err.Error())
		return
	}

	fmt.Printf("adding new directory: %s\n", newDir.Name)

	if err = c.AddDir(newDir.Path); err != nil {
		c.error(w, r, err.Error())
		return
	}

	fmt.Print("redirecting...")

	// Redirect to the page for the newly mapped out directory
	http.Redirect(w, r, homePage+"/dirs/i/"+newDir.ID, http.StatusOK)
}

func (c *Client) UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB
	file, handler, err := r.FormFile("file")
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	defer file.Close()

	// You can save the file on the server or process it as needed
	fmt.Fprintf(w, "Uploaded File: %+v\n", handler.Filename)
	fmt.Fprintf(w, "File Size: %+v\n", handler.Size)
	fmt.Fprintf(w, "MIME Header: %+v\n", handler.Header)
}

func (c *Client) UpdatePfpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // limit file size to 10mb
	file, handler, err := r.FormFile("profilePic")
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	defer file.Close()

	// Save the file to the client
	destPath, err := filepath.Abs("./assets/profile-pics")
	if err != nil {
		c.error(w, r, err.Error())
		return
	}
	filePath := filepath.Join(destPath, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("error: %v", err)
		c.error(w, r, err.Error())
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination and update client config accordingly
	if _, err := io.Copy(dst, file); err != nil {
		c.error(w, r, err.Error())
		return
	}
	if err := c.UpdateConfigSetting("CLIENT_PROFILE_PIC", filepath.Base(filePath)); err != nil {
		c.error(w, r, err.Error())
		return
	}

	// Redirect or respond with a success message
	http.Redirect(w, r, "/user", http.StatusOK)
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
	uploadPageData := UploadPage{
		UserPage:   userPage,
		ProfilePic: c.Conf.ProfilePic,
		ServerHost: c.Conf.ServerAddr,
		ClientHost: c.Conf.Addr,
	}
	err := c.Templates.ExecuteTemplate(w, "upload.html", uploadPageData)
	if err != nil {
		c.error(w, r, err.Error())
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
