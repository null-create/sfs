package db

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"

	svc "github.com/sfs/pkg/service"
)

func (q *Query) AddFile(f *svc.File) error {
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
		&f.ID,
		&f.Name,
		&f.OwnerID,
		&f.DirID,
		&f.DriveID,
		&f.Protected,
		&f.Key,
		&f.LastSync,
		&f.Path,
		&f.ServerPath,
		&f.ClientPath,
		&f.Endpoint,
		&f.CheckSum,
		&f.Algorithm,
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

	for _, f := range files {
		if err := q.Prepare(AddFileQuery); err != nil {
			return fmt.Errorf("failed to prepare statement: %v", err)
		}
		if _, err := q.Stmt.Exec(
			&f.ID,
			&f.Name,
			&f.OwnerID,
			&f.DirID,
			&f.DriveID,
			&f.Protected,
			&f.Key,
			&f.LastSync,
			&f.Path,
			&f.ServerPath,
			&f.ClientPath,
			&f.Endpoint,
			&f.CheckSum,
			&f.Algorithm,
		); err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}
		q.Stmt.Close()
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

	if q.Debug {
		log.Printf("[DEBUG] querying user: %s ", u.ID)
	}
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

func (q *Query) AddDir(d *svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddDirQuery); err != nil {
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

func (q *Query) AddDirs(dirs []*svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	for _, d := range dirs {
		if err := q.Prepare(AddDirQuery); err != nil {
			return fmt.Errorf("failed to prepare statement: %v", err)
		}
		if _, err := q.Stmt.Exec(
			&d.ID,
			&d.Name,
			&d.OwnerID,
			&d.DriveID,
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
		q.Stmt.Close()
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
		&drv.DriveRoot,
		&drv.RootID,
	); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}
