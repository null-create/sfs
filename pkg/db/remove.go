package db

import (
	"fmt"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

const (
	// remove a user iff they already exist in the database
	RemoveUserQuery string = "DELETE FROM Users WHERE id = '?' AND EXISTS (SELECT 1 FROM Users WHERE id = '?');"
	RemoveFileQuery string = "DELETE FROM Files WHERE id = '?' AND EXISTS (SELECT 1 FROM Users WHERE id = '?');"
)

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
