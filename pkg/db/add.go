package db

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/files"
)

// NOTE: these are string literals.
// may need to have all \t and \n characters
// removed before processing?
const (
	// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)
	AddFileQuery string = `
	INSERT OR IGNORE INTO Files (
		id, 
		name, 
		owner, 
		protected, 
		key, 
		path, 
		server_path, 
		client_path, 
		checksum, 
		algorithm
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	InsertUserQuery string = `
	INSERT OR IGNORE INTO Users (
		id, 
		name, 
		username, 
		email, 
		password, 
		last_login, 
		is_admin, 
		total_files, 
		total_directories
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`
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
		log.Fatal(err)
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
