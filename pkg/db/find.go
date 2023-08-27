package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

// TODO: handle null columns in all methods below

// TODO: test this!
// Query to check user existence
func (q *Query) UserExists(userID string) (bool, error) {
	var exists bool
	if err := q.Conn.QueryRow(ExistsQuery, "Users", userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("[ERROR] couldn't query user (%s): %v", userID, err)
	}
	return exists, nil
}

// get user data from database
func (q *Query) GetUser(userID string) (*auth.User, error) {
	q.Connect()
	defer q.Close()

	if q.Debug {
		log.Printf("[DEBUG] querying user %s", userID)
	}

	user := new(auth.User)
	if err := q.Conn.QueryRow(FindUserQuery, userID).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Password,
		&user.Email,
		&user.LastLogin,
		&user.Admin,
		&user.TotalFiles,
		&user.TotalDirs,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("[ERROR] no rows returned: %v", err)
		}
		return nil, fmt.Errorf("[ERROR] unable to execute query: %v", err)
	}

	return user, nil
}

// populate a slice of *auth.User structs from the database
func (q *Query) GetUsers() ([]*auth.User, error) {
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		var user *auth.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.UserName,
			&user.Email,
			&user.Password,
			&user.LastLogin,
			&user.Admin,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("[ERROR] no rows returned: %v", err)
			}
			return nil, fmt.Errorf("[ERROR] query failed: %v", err)
		}
		users = append(users, user)
	}
	if len(users) == 0 {
		log.Print("[DEBUG] no users found")
	}

	return users, nil
}

// retrieve file metadata from the database
func (q *Query) GetFile(fileID string) (*files.File, error) {
	q.Connect()
	defer q.Close()

	file := new(files.File)
	if err := q.Conn.QueryRow(FindFileQuery, fileID).Scan(
		&file.ID,
		&file.Name,
		&file.Owner,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned (id=%s): %v", fileID, err)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// populate a slice of *file.File structs from the database
//
// assign a limit with a string such as "100"
func (q *Query) GetFiles() ([]*files.File, error) {
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllQuery, "Files")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*files.File
	for rows.Next() {
		file := new(files.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.Owner,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[DEBUG] no rows returned: %v", err)
				continue
			}
			return nil, fmt.Errorf("[ERROR] unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	if len(fs) == 0 {
		log.Print("[DEBUG] no files returned")
	}
	return fs, nil
}

func (q *Query) GetDirectory(dirID string) (*files.Directory, error) {
	q.Connect()
	defer q.Close()

	d := new(files.Directory)
	if err := q.Conn.QueryRow(FindDirQuery, dirID).Scan(
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
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows found with dir id: %s", dirID)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] query failed: %v", err)
	}
	return d, nil
}

func (q *Query) GetDirectories(limit int) ([]*files.Directory, error) {
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllQuery, "Directories", limit)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*files.Directory, 0)
	for rows.Next() {
		dir := new(files.Directory)
		if err := q.Conn.QueryRow(FindAllQuery, "Directories").Scan(
			&dir.ID,
			&dir.Name,
			&dir.Owner,
			&dir.Size,
			&dir.Path,
			&dir.Protected,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[DEBUG] no rows returned: %v", err)
				continue
			}
			log.Fatalf("failed to get rows: %v", err)
		}
		dirs = append(dirs, dir)
	}
	if len(dirs) == 0 {
		log.Printf("[DEBUG] no directories returned")
	}
	return dirs, nil
}

func (q *Query) GetDrive(driveID string) (*files.Drive, error) {
	q.Connect()
	defer q.Close()

	d := new(files.Drive)
	if err := q.Conn.QueryRow(FindDriveQuery, driveID).Scan(
		&d.ID,
		&d.Name,
		&d.Owner,
		&d.TotalSize,
		&d.UsedSpace,
		&d.FreeSpace,
		&d.Protected,
		&d.Key,
		&d.AuthType,
		&d.DriveRoot,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] query failed: %v", err)
	}

	return d, nil
}

func (q *Query) GetDrives() ([]*files.Drive, error) { return nil, nil }
