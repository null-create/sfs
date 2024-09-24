package client

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sfs/pkg/env"

	"github.com/joeshaw/envdecode"
)

var EndpointRoot = "http://" + cfgs.Host

type Conf struct {
	IsAdmin         bool   `env:"ADMIN_MODE"`                   // whether the service should be run in admin mode or not
	BufferedEvents  bool   `env:"BUFFERED_EVENTS,required"`     // whether events should be buffered (i.e. have a delay between sync events)
	EventBufferSize int    `env:"EVENT_BUFFER_SIZE,required"`   // size of events buffer
	User            string `env:"CLIENT_NAME,required"`         // users name
	UserAlias       string `env:"CLIENT_USERNAME,required"`     // users alias (username)
	ID              string `env:"CLIENT_ID,required"`           // this is generated at creation time. won't be in the initial .env file
	Email           string `env:"CLIENT_EMAIL,required"`        // users email
	ProfilePic      string `env:"CLIENT_PROFILE_PIC,required"`  // path to users profile picture
	Root            string `env:"CLIENT_ROOT,required"`         // client service root (ie. ../sfs/client/run/)
	TestRoot        string `env:"CLIENT_TESTING,required"`      // testing root directory
	ClientPort      int    `env:"CLIENT_PORT,required"`         // client port
	Addr            string `env:"CLIENT_ADDRESS,required"`      // address for http client
	NewService      bool   `env:"CLIENT_NEW_SERVICE,required"`  // whether we need to initialize a new client service instance.
	LogDir          string `env:"CLIENT_LOG_DIR,required"`      // location of log directory
	LocalBackup     bool   `env:"CLIENT_LOCAL_BACKUP,required"` // whether we're backing up to a local file directory
	BackupDir       string `env:"CLIENT_BACKUP_DIR,required"`   // location of backup directory
	ServerAddr      string `env:"SERVER_ADDR,required"`         // server address
	Host            string `env:"SERVER_HOST,required"`         // client host
	Port            int    `env:"SERVER_PORT,required"`         // server port
	EnvFile         string `env:"SERVICE_ENV,required"`         // absoloute path to the dedicated .env file
}

func GetClientConfigs() *Conf {
	env.SetEnv(false)

	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("failed to decode client config .env file: %s", err)
	}
	return &c
}

// client env, user, and service configurations
var cfgs = GetClientConfigs()

// Environment configs
var envCfgs = env.NewE()

// update user and application configurations
func (c *Client) UpdateConfigSetting(setting, value string) error {
	switch setting {
	case "CLIENT_NAME":
		return c.updateClientName(value)
	case "CLIENT_USERNAME":
		return c.updateUserAlias(value)
	case "CLIENT_EMAIL":
		return c.updateClientEmail(value)
	case "CLIENT_PASSWORD":
		return c.updateUserPassword(c.User.Password, value)
	case "CLIENT_PORT":
		return c.updateClientPort(value)
	case "CLIENT_BACKUP_DIR":
		return c.UpdateBackupPath(value)
	case "CLIENT_LOCAL_BACKUP":
		return c.SetLocalBackup(value)
	case "CLIENT_PROFILE_PIC":
		return c.updateClientIcon(value)
	case "CLIENT_NOTIFICATIONS":
		fmt.Print("no implemented yet") // TODO:
	case "CLIENT_NEW_SERVICE":
		return c.updateClientNewService(value)
	case "NEW_SERVICE":
		return c.updateNewService(value)
	case "EVENT_BUFFER_SIZE":
		return c.updateEventBufferSize(value)
	default:
		return fmt.Errorf("unsupported setting: '%s'", setting)
	}
	return nil
}

// enable or disable backing up files to local storage.
func (c *Client) SetLocalBackup(modeStr string) error {
	mode, err := strconv.ParseBool(modeStr)
	if err != nil {
		c.log.Error(fmt.Sprintf("failed to parse string: %v", err))
		return err
	}
	c.Conf.LocalBackup = mode
	if err := envCfgs.Set("CLIENT_LOCAL_BACKUP", strconv.FormatBool(mode)); err != nil {
		return err
	}
	if mode {
		c.log.Info("local backup enabled")
	} else {
		c.log.Info("local backup disabled")
	}
	if err := c.SaveState(); err != nil {
		c.log.Error("failed to update state file: " + err.Error())
	}
	return nil
}

// update the backup configurations for the service and all client-side
// files and directories
func (c *Client) UpdateBackupPath(newDirPath string) error {
	if !c.isDirPath(newDirPath) {
		c.log.Error("path is not a directory")
		return fmt.Errorf("path is not a directory")
	}
	if c.LocalBackupDir == newDirPath {
		return nil // nothing to replace
	}
	return c.updateBackupPaths(filepath.Clean(newDirPath))
}

// If a user specifies a custom backup directory, then all
// local backup items need to have their backup paths updated.
// we don't want files to be backed up to a default directory if
// a user tries to specify otherwise.
func (c *Client) updateBackupPaths(newPath string) error {
	oldPath := c.Root
	dirs := c.Drive.GetDirs()
	for _, dir := range dirs {
		dir.BackupPath = strings.Replace(dir.BackupPath, oldPath, newPath, 1)
	}
	if err := c.Db.UpdateDirs(dirs); err != nil {
		return err
	}
	files := c.Drive.GetFiles()
	for _, file := range files {
		file.BackupPath = strings.Replace(file.BackupPath, oldPath, newPath, 1)
	}
	if err := c.Db.UpdateFiles(files); err != nil {
		return err
	}
	c.Conf.BackupDir = newPath
	if err := envCfgs.Set("CLIENT_BACKUP_DIR", newPath); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update user's name
func (c *Client) updateClientName(newName string) error {
	if newName == c.Conf.User && newName == c.User.Name && newName == c.Drive.OwnerName {
		return nil
	}
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %v", c.UserID)
	}
	c.User.Name = newName
	c.Conf.User = newName
	c.Drive.OwnerName = newName
	user.Name = newName
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := c.Db.UpdateDrive(c.Drive); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_NAME", newName); err != nil {
		return err
	}
	// TODO: sync with remote server, if necessary
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

func (c *Client) updateUserAlias(newAlias string) error {
	if newAlias == c.User.UserName || newAlias == "" {
		return nil
	}
	c.User.UserName = newAlias
	c.Conf.UserAlias = newAlias
	if err := c.Db.UpdateUser(c.User); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_USERNAME", newAlias); err != nil {
		return err
	}
	// TODO: sync with remote server, if necessary
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update the value for CLIENT_PROFILE_PIC. ususally a file name.
func (c *Client) updateClientIcon(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("no path specified")
	}
	c.Conf.ProfilePic = fileName
	if err := envCfgs.Set("CLIENT_PROFILE_PIC", fileName); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update user's email
func (c *Client) updateClientEmail(newEmail string) error {
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: '%s' (id=%v)", c.Conf.User, c.UserID)
	}
	c.User.Email = newEmail
	user.Email = newEmail
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_EMAIL", newEmail); err != nil {
		return err
	}
	// TODO: sync with remote server, if necessary
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update user's password
func (c *Client) updateUserPassword(oldPw, newPw string) error {
	user, err := c.Db.GetUser(c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: '%s' (id=%v)", c.Conf.User, c.UserID)
	}
	if newPw == user.Password && newPw == c.User.Password {
		return nil // nothing to update
	}
	// make sure current password is valid
	if oldPw != user.Password && oldPw != c.User.Password {
		return fmt.Errorf("incorrect password. password not updated")
	}
	user.Password = newPw
	c.User.Password = newPw
	// TODO: hashing of user passwords should occur before saving to DB.
	// dont save them as plaintext! or not save the PW at all?
	if err := c.Db.UpdateUser(user); err != nil {
		return err
	}
	if err := envCfgs.Set("CLIENT_PASSWORD", newPw); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// update client port setting
func (c *Client) updateClientPort(pvalue string) error {
	port, err := strconv.Atoi(pvalue)
	if err != nil {
		return err
	}
	if c.Conf.ClientPort == port {
		return nil // nothing to do here
	}
	c.Conf.ClientPort = port
	if err := envCfgs.Set("CLIENT_PORT", pvalue); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

func (c *Client) updateEventBufferSize(sizestr string) error {
	size, err := strconv.Atoi(sizestr)
	if err != nil {
		return err
	}
	if size == c.Conf.EventBufferSize {
		return nil
	}
	c.Conf.EventBufferSize = size
	if err := envCfgs.Set("EVENT_BUFFER_SIZE", sizestr); err != nil {
		return err
	}
	if err := c.SaveState(); err != nil {
		return err
	}
	return nil
}

// new service settings are strings that should be "true" or "false"
func (c *Client) validServiceSetting(value string) bool {
	return value == "true" || value == "false"
}

// resets new service to true for NEW_SERVICE and CLIENT_NEW_SERVICE env vars
func (c *Client) updateClientNewService(value string) error {
	if !c.validServiceSetting(value) {
		return fmt.Errorf("invalid value for new service setting (must be 'true' or 'false'): %v", value)
	}
	if err := envCfgs.Set("CLIENT_NEW_SERVICE", value); err != nil {
		return err
	}
	return nil
}

func (c *Client) updateNewService(value string) error {
	if !c.validServiceSetting(value) {
		return fmt.Errorf("invalid value for new service setting (must be 'true' or 'false'): %v", value)
	}
	if err := envCfgs.Set("NEW_SERVICE", value); err != nil {
		return err
	}
	return nil
}
