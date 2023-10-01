package db

import (
	"fmt"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

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

// remove a table and create a new one by the same name
func (q *Query) ResetTable(tableName string) error {
	if err := q.DropTable(tableName); err != nil {
		return err
	}
	tablePath := filepath.Join(q.DBPath, tableName)
	if err := NewDB(tableName, tablePath); err != nil {
		return err
	}
	return nil
}

func (q *Query) RemoveUser(userID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveUserQuery, userID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user: %v", err)
	}
	return nil
}

func (q *Query) RemoveUsers(users []*auth.User) error {
	q.Connect()
	defer q.Close()

	for _, user := range users {
		_, err := q.Conn.Exec(RemoveUserQuery, user.ID, user.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user: %v", err)
		}
	}
	return nil
}

func (q *Query) RemoveFile(fileID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveFileQuery, fileID, fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file (id=%s): %v", fileID, err)
	}
	return nil
}

func (q *Query) RemoveFiles(fs []*svc.File) error {
	q.Connect()
	defer q.Close()

	for _, f := range fs {
		_, err := q.Conn.Exec(RemoveFileQuery, f.ID, f.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user: %v", err)
		}
	}
	return nil
}

func (q *Query) RemoveDirectory(dirID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveDirectoryQuery, dirID, dirID)
	if err != nil {
		return fmt.Errorf("failed to remove file (id=%s): %v", dirID, err)
	}
	return nil
}

func (q *Query) RemoveDirectories(dirs []*svc.Directory) error {
	q.Connect()
	defer q.Close()

	for _, d := range dirs {
		_, err := q.Conn.Exec(RemoveDirectoryQuery, d.ID)
		if err != nil {
			return fmt.Errorf("failed to remove file (id=%s): %v", d.ID, err)
		}
	}
	return nil
}

func (q *Query) RemoveDrive(driveID string) error {
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveDriveQuery, driveID, driveID)
	if err != nil {
		return fmt.Errorf("failed to remove drive (id=%s): %v", driveID, err)
	}
	return nil
}

func (q *Query) RemoveDrives(drvs []*svc.Drive) error {
	q.Connect()
	defer q.Close()

	for _, drv := range drvs {
		_, err := q.Conn.Exec(RemoveDriveQuery, drv.ID, drv.ID)
		if err != nil {
			return fmt.Errorf("failed to remove file (id=%s): %v", drv.ID, err)
		}
	}
	return nil
}
