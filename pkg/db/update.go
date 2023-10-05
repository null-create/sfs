package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

func (q *Query) UpdateFile(f *svc.File) error {
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateFileQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&f.ID,
		&f.Name,
		&f.Owner,
		&f.Protected,
		&f.Key,
		&f.LastSync,
		&f.Path,
		&f.ServerPath,
		&f.ClientPath,
		&f.CheckSum,
		&f.Algorithm,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) UpdateDir(d *svc.Directory) error {
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateDirQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&d.ID,
		&d.Name,
		&d.Owner,
		&d.Size,
		&d.Path,
		&d.Protected,
		&d.AuthType,
		&d.Key,
		&d.Overwrite,
		&d.LastSync,
		&d.Root,
		&d.RootPath,
	); err != nil {
		return fmt.Errorf("failed to add directory: %v", err)
	}
	return nil
}

func (q *Query) UpdateDrive(drv *svc.Drive) error {
	q.Connect()
	defer q.Close()

	if err := q.Prepare(UpdateDriveQuery); err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	if _, err := q.Stmt.Exec(
		&drv.ID,
		&drv.Name,
		&drv.Owner,
		&drv.TotalSize,
		&drv.UsedSpace,
		&drv.FreeSpace,
		&drv.Protected,
		&drv.Key,
		&drv.AuthType,
		&drv.DriveRoot,
	); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}

func (q *Query) UpdateUser(u *auth.User) error {
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
		&u.Root,
		&u.ID,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}
	return nil
}
