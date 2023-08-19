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
		&f.LastSync,
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

// iterate over the files and execute the statement for each
func (q *Query) AddFiles(fs []*files.File) error {
	for _, file := range fs {
		if err := q.AddFile(file); err != nil {
			log.Fatalf("[ERROR] failed to add file: %v", err)
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

// add a slice of users to the users database
func (q *Query) AddUsers(usrs []*auth.User) error {
	for _, u := range usrs {
		if err := q.AddUser(u); err != nil {
			log.Fatalf("[ERROR] failed to add user: %v", err)
		}
	}
	return nil
}

func (q *Query) AddDir(d *files.Directory) error {

	return nil
}

func (q *Query) AddDirs(dirs []*files.Directory) error {

	return nil
}

func (q *Query) AddDrive(drv *files.Drive) error {

	return nil
}

func (q *Query) AddDrives(drvs []*files.Drive) error {

	return nil
}
