package db

import (
	"database/sql"
	"fmt"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/files"
)

// TODO: handle null columns in all methods below

// Query to check user existence
func (q *Query) UserExists(userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM Users WHERE id = ?)`

	var exists bool
	if err := q.Conn.QueryRow(query, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("[ERROR] couldn't query user (%s): %v", userID, err)
	}
	return exists, nil
}

// get user data from database
func (q *Query) GetUser(userID string) (*auth.User, error) {
	q.Connect()
	defer q.Close()

	var user *auth.User
	if err := q.Conn.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Password,
		&user.Email,
		&user.LastLogin,
		user.IsAdmin,
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

	rows, err := q.Conn.Query("SELECT * FROM users;")
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
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// retrieve file metadata from the database
func (q *Query) GetFile(fileID string) (*files.File, error) {
	q.Connect()
	defer q.Close()

	var file *files.File
	if err := q.Conn.QueryRow("SELECT * FROM files WHERE id = ?;", fileID).Scan(
		&file.ID,
		&file.Name,
		&file.Owner,
		&file.Protected,
		&file.Key,
		&file.Path,
		&file.ClientPath,
		&file.ServerPath,
		&file.LastSync,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("[ERROR] no rows returned: %v", err)
		}
		return nil, fmt.Errorf("[ERROR] failed to get file metadata: %v", err)
	}
	return file, nil
}

// populate a slice of *file.File structs from the database
//
// assign a limit with a string such as "100"
func (q *Query) GetFiles(limit string) ([]*files.File, error) {
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query("SELECT * FROM files LIMIT ?;", limit)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*files.File
	for rows.Next() {
		var file *files.File
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
				return nil, fmt.Errorf("[ERROR] no rows returned: %v", err)
			}
			return nil, fmt.Errorf("[ERROR] unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[ERROR] %v", err)
	}

	return fs, nil
}

func (q *Query) GetDirectory(dirID string) (*files.Directory, error) {
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query("SELECT * FROM directories LIMIT ?;")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	return nil, nil
}
