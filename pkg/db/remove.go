package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// remove a table from the database
func Drop(dbPath string, tableName string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(DropTableQuery, tableName)
	if err != nil {
		log.Fatalf("[ERROR] failed to drop table: \n%v\n", err)
	}
}

// delete a table if it exists
func (q *Query) DropTable(tableName string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(DropTableQuery, tableName)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %v", tableName, err)
	}

	return nil
}

func (q *Query) RemoveUser(userID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveQuery, "Users", userID)
	if err != nil {
		return fmt.Errorf("failed to remove user: %v", err)
	}

	return nil
}

func (q *Query) RemoveUsers(u []*auth.User) error {
	q.Connect()
	defer q.Close()

	for _, user := range u {
		_, err := q.Conn.Exec(RemoveQuery, "Users", user.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user: %v", err)
		}
	}

	return nil
}

func (q *Query) RemoveFile(fileID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveQuery, "Files", fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file (id=%s): %v", fileID, err)
	}

	return nil
}

func (q *Query) RemoveFiles(fs []*svc.File) error {
	q.Connect()
	defer q.Close()

	for _, f := range fs {
		_, err := q.Conn.Exec(RemoveQuery, "Files", f.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user: %v", err)
		}
	}

	return nil
}

func (q *Query) RemoveDirectory(dirID string) error { return nil }

func (q *Query) RemoveDirectories(d []*svc.Directory) error { return nil }
