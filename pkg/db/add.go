package db

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

func (q *Query) AddFile(f *files.File) error {
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

// iterate over the files and execute the statement for each
func (q *Query) AddFiles(fs []*files.File) error {
	for _, file := range fs {
		if err := q.AddFile(file); err != nil {
			return fmt.Errorf("[ERROR] : %v", err)
		}
	}
	return nil
}

// add a user to the user database
func (q *Query) AddUser(u *auth.User) error {
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
		&u.Password,
		&u.Email,
		&u.LastLogin,
		&u.Admin,
		&u.TotalFiles,
		&u.TotalDirs,
	); err != nil {
		return fmt.Errorf("failed to execute statement: %v", err)
	}

	return nil
}

// add a slice of users to the users database
func (q *Query) AddUsers(usrs []*auth.User) error {
	for _, u := range usrs {
		if err := q.AddUser(u); err != nil {
			return fmt.Errorf("[ERROR] failed to add user: %v", err)
		}
	}
	return nil
}

func (q *Query) AddDir(d *files.Directory) error {
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

func (q *Query) AddDirs(dirs []*files.Directory) error {
	for _, dir := range dirs {
		if err := q.AddDir(dir); err != nil {
			return fmt.Errorf("[ERROR] failed to add dir (%s) to database: %v", dir.ID, err)
		}
	}
	return nil
}

func (q *Query) AddDrive(drv *files.Drive) error {
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddDriveQuery); err != nil {
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

func (q *Query) AddDrives(drvs []*files.Drive) error {
	if len(drvs) == 0 {
		return fmt.Errorf("no drives inputted")
	}
	for _, d := range drvs {
		if err := q.AddDrive(d); err != nil {
			return fmt.Errorf("failed to add drive: %v", err)
		}
	}
	return nil
}
