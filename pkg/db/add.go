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
		return fmt.Errorf("[ERROR] failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	// execute the statement
	if _, err := q.Stmt.Exec(
		&f.ID,
		&f.Name,
		&f.Owner,
		&f.Protected,
		&f.Key,
		&f.Path,
		&f.ServerPath,
		&f.ClientPath,
		&f.CheckSum,
		&f.Algorithm,
	); err != nil {
		return fmt.Errorf("[ERROR] failed to execute statement: %v", err)
	}
	return nil
}

func (q *Query) AddFiles(fs *[]files.File) error {
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddFileQuery); err != nil {
		return fmt.Errorf("[ERROR] failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	// iterate over the files and execute the statement for each
	for _, file := range *fs {
		if _, err := q.Stmt.Exec(
			&file.ID,
			&file.Name,
			&file.Owner,
			&file.Protected,
			&file.Key,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			log.Fatalf("[ERROR] unable to execute statement: %v", err)
		}
	}

	return nil
}

func (q *Query) AddUser(u *auth.User) error {
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddUserQuery); err != nil {
		return fmt.Errorf("[ERROR] failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

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
		return fmt.Errorf("[ERROR] failed to execute statement: %v", err)
	}

	return nil
}

func (q *Query) AddUsers(usrs []*auth.User) error {
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(AddUserQuery); err != nil {
		return fmt.Errorf("[ERROR] failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	for _, u := range usrs {
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
			return fmt.Errorf("[ERROR] failed to execute statement: %v", err)
		}
	}

	return nil
}

func (q *Query) AddDir(d *files.Directory) error { return nil }

func (q *Query) AddDirs(dirs []*files.Directory) error { return nil }

func (q *Query) AddDrive(drv *files.Drive) error { return nil }

func (q *Query) AddDrives(drvs []*files.Drive) error { return nil }
