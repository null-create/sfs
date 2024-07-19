package db

import (
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// attempts ot map a given db to its core table.
// returns an empty string if none can be matched.
func (q *Query) getTable(dbName string) string {
	if dbName == "" {
		log.Printf("[WARNING] no database provided")
		return ""
	}
	switch dbName {
	case "users":
		return "Users"
	case "drives":
		return "Drives"
	case "directories":
		return "Directories"
	case "files":
		return "Files"
	}
	return ""
}

func (q *Query) getResetQueries(tableName string) (string, string) {
	var dropQuery string
	var createQuery string
	switch tableName {
	case "Users":
		dropQuery = DropUserTableQuery
		createQuery = CreateUserTable
	case "Drives":
		dropQuery = DropDrivesTableQuery
		createQuery = CreateDriveTable
	case "Directories":
		dropQuery = DropDirectoriesTableQuery
		createQuery = CreateDirectoryTable
	case "Files":
		dropQuery = DropFilesTableQuery
		createQuery = CreateFileTable
	default:
		log.Fatalf("unsupported table name: %s", tableName)
	}
	return dropQuery, createQuery
}

// delete a table if it exists
func (q *Query) DropTable(dbName string) error {
	q.WhichDB(dbName)
	q.Connect()
	defer q.Close()

	var query string
	switch dbName {
	case "users":
		query = DropUserTableQuery
	case "drives":
		query = DropDrivesTableQuery
	case "directories":
		query = DropDirectoriesTableQuery
	case "files":
		query = DropFilesTableQuery
	}
	_, err := q.Conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}

func (q *Query) RemoveUser(userID string) error {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveUserQuery, userID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user: %v", err)
	}
	return nil
}

func (q *Query) RemoveUsers(users []*auth.User) error {
	q.WhichDB("users")
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
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveFileQuery, fileID, fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file (id=%s): %v", fileID, err)
	}
	return nil
}

func (q *Query) RemoveFiles(files []*svc.File) error {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	for _, file := range files {
		_, err := q.Conn.Exec(RemoveFileQuery, file.ID, file.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user: %v", err)
		}
	}
	return nil
}

func (q *Query) RemoveDirectory(dirID string) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveDirectoryQuery, dirID, dirID)
	if err != nil {
		return fmt.Errorf("failed to remove file (id=%s): %v", dirID, err)
	}
	return nil
}

func (q *Query) RemoveDirectories(dirs []*svc.Directory) error {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	for _, dir := range dirs {
		_, err := q.Conn.Exec(RemoveDirectoryQuery, dir.ID)
		if err != nil {
			return fmt.Errorf("failed to remove file (id=%s): %v", dir.ID, err)
		}
	}
	return nil
}

func (q *Query) RemoveDrive(driveID string) error {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	_, err := q.Conn.Exec(RemoveDriveQuery, driveID, driveID)
	if err != nil {
		return fmt.Errorf("failed to remove drive (id=%s) from database: %v", driveID, err)
	}
	return nil
}

func (q *Query) RemoveDrives(drvs []*svc.Drive) error {
	q.WhichDB("drives")
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

// "clears" a database by dropping the associated table for the given
// database name and recreates it entirely.
func (q *Query) ClearTable(dbName string) error {
	q.WhichDB(dbName)
	q.Connect()
	defer q.Close()

	tableName := q.getTable(dbName)
	if tableName == "" {
		return fmt.Errorf("no table found for given DB: %s", dbName)
	}
	// drop the table
	dropQuery, createQuery := q.getResetQueries(tableName)
	_, err := q.Conn.Exec(dropQuery)
	if err != nil {
		return err
	}
	// re-create the table
	_, err = q.Conn.Exec(createQuery)
	if err != nil {
		return err
	}
	return nil
}
