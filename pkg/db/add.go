package db

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/files"
)

// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)
const query = `
INSERT OR IGNORE INTO Files (id, name, owner, protected, key, path, server_path, client_path, checksum, algorithm)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

func (q *Query) AddFile(f *files.File) error {
	q.Connect()
	defer q.Close()

	// prepare query
	stmt, err := q.Conn.Prepare(query)
	if err != nil {
		log.Fatalf("[ERROR] failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// execute the statement
	if _, err = stmt.Exec(
		f.ID,
		f.Name,
		f.Owner,
		f.Protected,
		f.Key,
		f.Path,
		f.ServerPath,
		f.ClientPath,
		f.CheckSum,
		f.Algorithm,
	); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (q *Query) AddFiles(f *[]files.File) error {
	q.Connect()
	defer q.Close()

	// prepare query
	if err := q.Prepare(query); err != nil {
		return fmt.Errorf("[ERROR] failed to prepare statement: %v", err)
	}
	defer q.Stmt.Close()

	// iterate over the files and execute the statement for each
	for _, file := range *f {
		if _, err := q.Stmt.Exec(
			file.ID,
			file.Name,
			file.Owner,
			file.Protected,
			file.Key,
			file.Path,
			file.ServerPath,
			file.ClientPath,
			file.CheckSum,
			file.Algorithm,
		); err != nil {
			log.Fatalf("[ERROR] unable to execute statement: %v", err)
		}
	}

	return nil
}
