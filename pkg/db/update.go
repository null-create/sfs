package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

func (q *Query) UpdateFile(file *svc.File) error {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		UpdateFileQuery,
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Mode,
		&file.Size,
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
		&file.ID,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) UpdateFiles(files []*svc.File) error {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	for _, file := range files {
		if _, err := q.Conn.Exec(
			UpdateFileQuery,
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
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
			&file.ID,
		); err != nil {
			return err
		}
	}
	return nil
}

func (q *Query) UpdateDir(dir *svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		UpdateDirQuery,
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
		&dir.ID,
	); err != nil {
		return fmt.Errorf("failed to add directory: %v", err)
	}
	return nil
}

func (q *Query) UpdateDirs(dirs []*svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	for _, dir := range dirs {
		if _, err := q.Conn.Exec(
			UpdateDirQuery,
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
			&dir.ID,
		); err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}
	}
	return nil
}

func (q *Query) UpdateDrive(drv *svc.Drive) error {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		UpdateDriveQuery,
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
		&drv.ID,
	); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}

func (q *Query) UpdateUser(user *auth.User) error {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	if _, err := q.Conn.Exec(
		UpdateUserQuery,
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
		&user.ID,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}
