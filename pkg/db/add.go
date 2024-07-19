package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"

	svc "github.com/sfs/pkg/service"
)

func (q *Query) AddFile(file *svc.File) error {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddFileQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	// execute the statement
	if _, err := q.Stmt.Exec(
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
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(AddFileQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	for _, file := range files {
		if _, err := q.Stmt.Exec(
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
func (q *Query) AddUser(u *auth.User) error {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddUserQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&u.ID,
		&u.Name,
		&u.UserName,
		&u.Email,
		&u.Password,
		&u.LastLogin,
		&u.Admin,
		&u.SfPath,
		&u.DriveID,
		&u.TotalFiles,
		&u.TotalDirs,
		&u.DrvRoot,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) AddDir(dir *svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddDirQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
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
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(AddDirQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	for _, dir := range dirs {
		if _, err := q.Stmt.Exec(
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
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddDriveQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
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
