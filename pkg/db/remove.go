package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

const (
	// drop table query
	DropQuery string = "DROP TABLE IF EXISTS ?;"

	// remove a user or file iff they (or the file) already exists in the database
	RemoveUserQuery string = "DELETE FROM Users WHERE id = '?' AND EXISTS (SELECT 1 FROM Users WHERE id = '?');"
	RemoveFileQuery string = "DELETE FROM Files WHERE id = '?' AND EXISTS (SELECT 1 FROM Users WHERE id = '?');"
)

// remove a table from the database
func drop(dbPath string, tableName string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(DropQuery, tableName)
	if err != nil {
		log.Fatalf("[ERROR] failed to drop table: \n%v\n", err)
	}
}

// delete a table if it exists
func (q *Query) DropTable(tableName string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(DropQuery, tableName)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to drop table %s: %v", tableName, err)
	}

	return nil
}

func (q *Query) RemoveUser(userID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveUserQuery, userID)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to remove user: %v", err)
	}

	return nil
}

func (q *Query) RemoveUsers(u []*auth.User) error {
	q.Connect()
	defer q.Close()

	for _, user := range u {
		_, err := q.Conn.Exec(RemoveUserQuery, user.ID)
		if err != nil {
			return fmt.Errorf("[ERROR] failed to remove user: %v", err)
		}
	}

	return nil
}

func (q *Query) RemoveFile(fileID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveFileQuery, fileID)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to remove file (id=%s): %v", fileID, err)
	}

	return nil
}

func (q *Query) RemoveFiles(fs []*files.File) error {
	q.Connect()
	defer q.Close()

	for _, f := range fs {
		_, err := q.Conn.Exec(RemoveFileQuery, f.ID)
		if err != nil {
			return fmt.Errorf("[ERROR] failed to remove user: %v", err)
		}
	}

	return nil
}

func (q *Query) RemoveDirectory(dirID string) error { return nil }

func (q *Query) RemoveDirectories(d []*files.Directory) error { return nil }
