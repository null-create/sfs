package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

func (q *Query) UpdateFile(f *svc.File) error {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateFileQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&f.ID,
		&f.Name,
		&f.OwnerID,
		&f.DirID,
		&f.DriveID,
		&f.Mode,
		&f.Size,
		&f.Backup,
		&f.Protected,
		&f.Key,
		&f.LastSync,
		&f.Path,
		&f.ServerPath,
		&f.ClientPath,
		&f.Endpoint,
		&f.CheckSum,
		&f.Algorithm,
		&f.ID,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) UpdateDir(d *svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateDirQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&d.ID,
		&d.Name,
		&d.OwnerID,
		&d.DriveID,
		&d.Size,
		&d.Path,
		&d.ServerPath,
		&d.ClientPath,
		&d.Protected,
		&d.AuthType,
		&d.Key,
		&d.Overwrite,
		&d.LastSync,
		&d.Endpoint,
		&d.Root,
		&d.RootPath,
		&d.ID,
	); err != nil {
		return fmt.Errorf("failed to add directory: %v", err)
	}
	return nil
}

func (q *Query) UpdateDrive(drv *svc.Drive) error {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateDriveQuery); err != nil {
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
		&drv.ID,
	); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}

func (q *Query) UpdateUser(u *auth.User) error {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateUserQuery); err != nil {
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
		&u.ID,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}
