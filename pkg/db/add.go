package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"

	svc "github.com/sfs/pkg/service"
)

func (q *Query) AddFile(file *svc.File) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	// execute the statement
	if _, err := q.Conn.Exec(
		AddFileQuery,
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Mode,
		&file.Size,
		&file.LocalBackup,
		&file.ServerBackup,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.BackupPath,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

// iterate over the files and execute the statement for each
func (q *Query) AddFiles(files []*svc.File) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	for _, file := range files {
		if _, err := q.Conn.Exec(
			AddFileQuery,
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}
	}
	return nil
}

// add a user to the user database
func (q *Query) AddUser(user *auth.User) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		AddUserQuery,
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.LastLogin,
		&user.Admin,
		&user.SfPath,
		&user.DriveID,
		&user.TotalFiles,
		&user.TotalDirs,
		&user.DrvRoot,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) AddDir(dir *svc.Directory) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		AddDirQuery,
		&dir.ID,
		&dir.Name,
		&dir.OwnerID,
		&dir.DriveID,
		&dir.Size,
		&dir.Path,
		&dir.ServerPath,
		&dir.ClientPath,
		&dir.BackupPath,
		&dir.Protected,
		&dir.AuthType,
		&dir.Key,
		&dir.Overwrite,
		&dir.LastSync,
		&dir.Endpoint,
		&dir.Root,
		&dir.RootPath,
	); err != nil {
		return fmt.Errorf("failed to add directory: %v", err)
	}
	return nil
}

func (q *Query) AddDirs(dirs []*svc.Directory) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	for _, dir := range dirs {
		if _, err := q.Conn.Exec(
			AddDirQuery,
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
			&dir.DriveID,
			&dir.Size,
			&dir.Path,
			&dir.ServerPath,
			&dir.ClientPath,
			&dir.BackupPath,
			&dir.Protected,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Endpoint,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			return fmt.Errorf("failed to add directory: %v", err)
		}
	}
	return nil
}

// add drive info to drive database
func (q *Query) AddDrive(drv *svc.Drive) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		AddDriveQuery,
		&drv.ID,
		&drv.OwnerName,
		&drv.OwnerID,
		&drv.TotalSize,
		&drv.UsedSpace,
		&drv.FreeSpace,
		&drv.Protected,
		&drv.Key,
		&drv.AuthType,
		&drv.IsLoaded,
		&drv.RootPath,
		&drv.RootID,
		&drv.Registered,
		&drv.RecycleBin,
	); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}
