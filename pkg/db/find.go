package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/logger"
	svc "github.com/sfs/pkg/service"
)

// Query to check user existence
func (q *Query) UserExists(userID string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	var exists bool
	err := q.Conn.QueryRow(ExistsQuery, "Users", userID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to query user %s: %v", userID, err)
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists, nil
}

// get user data from database. returns nil if user is not found.
func (q *Query) GetUser(userID string) (*auth.User, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	if q.Debug {
		q.log.Info(fmt.Sprintf("querying user %s", userID))
	}

	user := new(auth.User)
	if err := q.Conn.QueryRow(FindUserQuery, userID).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.LastLogin,
		&user.Admin,
		&user.SfPath,
		&user.DriveID,
		&user.TotalFiles,
		&user.TotalDirs,
		&user.DrvRoot,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log("INFO", fmt.Sprintf("no rows returned: %v", err))
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] unable to execute query: %v", err)
	}
	return user, nil
}

// get a userID from a driveID.
// will return an empty string if no userID is found with this driveID
func (q *Query) GetUserIDFromDriveID(driveID string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	if q.Debug {
		q.log.Info(fmt.Sprintf("querying for userID using driveID %s", driveID))
	}

	var userID string
	if err := q.Conn.QueryRow(FindUsersIDWithDriveIDQuery, driveID).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no user associated with driveID %s", driveID))
			return "", nil
		} else {
			return "", fmt.Errorf("failed to query for userID: %v", err)
		}
	}
	return userID, nil
}

// populates a slice of *auth.User structs of all available users from the database
func (q *Query) GetUsers() ([]*auth.User, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		user := new(auth.User)
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.UserName,
			&user.Email,
			&user.Password,
			&user.LastLogin,
			&user.Admin,
			&user.SfPath,
			&user.DriveID,
			&user.TotalFiles,
			&user.TotalDirs,
			&user.DrvRoot,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "users found in database")
				return nil, nil
			}
			return nil, fmt.Errorf("query failed: %v", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// ----- files ----------------------------------

// retrieve file metadata from the database
//
// file returns nil if no result is available
func (q *Query) GetFileByID(fileID string) (*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	file := new(svc.File)
	if err := q.Conn.QueryRow(FindFileQuery, fileID).Scan(
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Mode,
		&file.Size,
		&file.LocalBackup,
		&file.ServerBackup,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.BackupPath,
		&file.Registered,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows returned (id=%s): %v", fileID, err))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// find a file in the database by searching with its path
// returns nil if no file is found.
func (q *Query) GetFileByPath(filePath string) (*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	file := new(svc.File)
	if err := q.Conn.QueryRow(FindFileByPathQuery, filePath).Scan(
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Mode,
		&file.Size,
		&file.LocalBackup,
		&file.ServerBackup,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.BackupPath,
		&file.Registered,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows returned for file (path=%s): %v", filePath, err))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// get a file by name. returns nil if no file is found in the db.
func (q *Query) GetFileByName(fileName string) (*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	file := new(svc.File)
	if err := q.Conn.QueryRow(FindFilesByNameQuery, fileName).Scan(
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Mode,
		&file.Size,
		&file.LocalBackup,
		&file.ServerBackup,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.BackupPath,
		&file.Registered,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows returned (file name=%s): %v", fileName, err))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// populate a slice of *svc.File objects that all have the same name
func (q *Query) GetFilesByName(fileName string) ([]*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindFilesByNameQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*svc.File
	for rows.Next() {
		file := new(svc.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Registered,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "files found in database")
				continue
			}
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	return fs, nil
}

// retrieves a file ID using a given file path
func (q *Query) GetFileIDFromPath(filePath string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	var fileID string
	if err := q.Conn.QueryRow(FindFileIDWithPathQuery, filePath).Scan(&fileID); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows returned (path=%s): %v", filePath, err))
			return "", nil
		}
		return "", fmt.Errorf("failed to get file ID: %v", err)
	}
	return fileID, nil
}

func (q *Query) IsFileRegistered(fileID string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	var isRegistered bool
	if err := q.Conn.QueryRow(IsFileRegisteredQuery, fileID).Scan(&isRegistered); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return isRegistered, nil
}

// populate a slice of *svc.File structs from *all*
// files in the files database
func (q *Query) GetFiles() ([]*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllFilesQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*svc.File
	for rows.Next() {
		file := new(svc.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Registered,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "files found in database")
				continue
			}
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	return fs, nil
}

// populate a slice of *svc.File structs for all files
// associated with the given user. will return an empty slice
// if no files are found.
func (q *Query) GetUsersFiles(userID string) ([]*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersFilesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*svc.File
	for rows.Next() {
		file := new(svc.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Registered,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, fmt.Sprintf("files found for user (id=%s)", userID))
				continue
			}
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	return fs, nil
}

func (q *Query) GetFilesByDirID(dirID string) ([]*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindFilesByDirIDQuery, dirID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*svc.File
	for rows.Next() {
		file := new(svc.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Registered,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, fmt.Sprintf("files found with parent dir '%s'", dirID))
				continue
			}
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	return fs, nil
}

func (q *Query) GetFilesByDriveID(driveID string) ([]*svc.File, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindFilesByDriveIDQuery, driveID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	var fs []*svc.File
	for rows.Next() {
		file := new(svc.File)
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.OwnerID,
			&file.DirID,
			&file.DriveID,
			&file.Mode,
			&file.Size,
			&file.LocalBackup,
			&file.ServerBackup,
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.BackupPath,
			&file.Registered,
			&file.Endpoint,
			&file.CheckSum,
			&file.Algorithm,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, fmt.Sprintf("files found for user (id=%s)", driveID))
				continue
			}
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
		fs = append(fs, file)
	}
	return fs, nil
}

// ----------- directories --------------------------------

// retrieve information about a users directory from the database
//
// returns nil if no information is available
func (q *Query) GetDirectoryByID(dirID string) (*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	dir := new(svc.Directory)
	dir.Files = make(map[string]*svc.File, 0)
	dir.Dirs = make(map[string]*svc.Directory, 0)

	if err := q.Conn.QueryRow(FindDirQuery, dirID).Scan(
		&dir.ID,
		&dir.Name,
		&dir.OwnerID,
		&dir.DriveID,
		&dir.Size,
		&dir.Path,
		&dir.ServerPath,
		&dir.ClientPath,
		&dir.BackupPath,
		&dir.Registered,
		&dir.Protected,
		&dir.AuthType,
		&dir.Key,
		&dir.Overwrite,
		&dir.LastSync,
		&dir.Endpoint,
		&dir.ParentID,
		&dir.Root,
		&dir.RootPath,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows found with dir id: %s", dirID))
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return dir, nil
}

// find a directory by name. returns nil if no directory is found.
func (q *Query) GetDirectoryByName(dirName string) (*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	dir := new(svc.Directory)
	dir.Files = make(map[string]*svc.File, 0)
	dir.Dirs = make(map[string]*svc.Directory, 0)

	if err := q.Conn.QueryRow(FindDirByNameQuery, dirName).Scan(
		&dir.ID,
		&dir.Name,
		&dir.OwnerID,
		&dir.DriveID,
		&dir.Size,
		&dir.Path,
		&dir.ServerPath,
		&dir.ClientPath,
		&dir.BackupPath,
		&dir.Registered,
		&dir.Protected,
		&dir.AuthType,
		&dir.Key,
		&dir.Overwrite,
		&dir.LastSync,
		&dir.Endpoint,
		&dir.ParentID,
		&dir.Root,
		&dir.RootPath,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows found with dir name: %s", dirName))
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return dir, nil
}

// find a directory by name. returns nil if no directory is found.
func (q *Query) GetDirectoryByPath(dirPath string) (*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	dir := new(svc.Directory)
	dir.Files = make(map[string]*svc.File, 0)
	dir.Dirs = make(map[string]*svc.Directory, 0)

	if err := q.Conn.QueryRow(FindDirByPathQuery, dirPath).Scan(
		&dir.ID,
		&dir.Name,
		&dir.OwnerID,
		&dir.DriveID,
		&dir.Size,
		&dir.Path,
		&dir.ServerPath,
		&dir.ClientPath,
		&dir.BackupPath,
		&dir.Registered,
		&dir.Protected,
		&dir.AuthType,
		&dir.Key,
		&dir.Overwrite,
		&dir.LastSync,
		&dir.Endpoint,
		&dir.ParentID,
		&dir.Root,
		&dir.RootPath,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no rows found for dir: %s", filepath.Base(dirPath)))
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return dir, nil
}

// get a directory's ID from its absolute path.
func (q *Query) GetDirIDFromPath(dirPath string) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	var id string
	if err := q.Conn.QueryRow(FindDirIDByPathQuery, dirPath).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, fmt.Sprintf("no dir ID found from path: %v", dirPath))
			return "", nil
		}
		return "", fmt.Errorf("failed to get dir id: %v", err)
	}
	return id, nil
}

// geta a slice of *all* directories on the server, regardless
// of owner. only for admin users.
func (q *Query) GetAllDirectories() ([]*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllDirsQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*svc.Directory, 0)
	for rows.Next() {
		dir := new(svc.Directory)
		dir.Files = make(map[string]*svc.File, 0)
		dir.Dirs = make(map[string]*svc.Directory, 0)
		if err := rows.Scan(
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
			&dir.DriveID,
			&dir.Size,
			&dir.Path,
			&dir.ServerPath,
			&dir.ClientPath,
			&dir.BackupPath,
			&dir.Registered,
			&dir.Protected,
			&dir.Endpoint,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Endpoint,
			&dir.ParentID,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "no rows returned")
				continue
			}
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

// get all the directories for this user.
func (q *Query) GetUsersDirectories(userID string) ([]*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersDirectoriesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*svc.Directory, 0)
	for rows.Next() {
		dir := new(svc.Directory)
		dir.Files = make(map[string]*svc.File, 0)
		dir.Dirs = make(map[string]*svc.Directory, 0)
		if err := rows.Scan(
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
			&dir.DriveID,
			&dir.Size,
			&dir.Path,
			&dir.ServerPath,
			&dir.ClientPath,
			&dir.BackupPath,
			&dir.Registered,
			&dir.Protected,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Endpoint,
			&dir.ParentID,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "no rows returned")
				continue
			}
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

// get all the directories for this user.
func (q *Query) GetDirsByDriveID(driveID string) ([]*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindDirsByDriveIDQuery, driveID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*svc.Directory, 0)
	for rows.Next() {
		dir := new(svc.Directory)
		dir.Files = make(map[string]*svc.File, 0)
		dir.Dirs = make(map[string]*svc.Directory, 0)
		if err := rows.Scan(
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
			&dir.DriveID,
			&dir.Size,
			&dir.Path,
			&dir.ServerPath,
			&dir.ClientPath,
			&dir.BackupPath,
			&dir.Registered,
			&dir.Protected,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Endpoint,
			&dir.ParentID,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "no rows returned")
				continue
			}
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

// get all the directories with this parent id. returns an empty slice if none are found.
func (q *Query) GetDirsByParentID(parentID string) ([]*svc.Directory, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindDirsByParentIDQuery, parentID)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*svc.Directory, 0)
	for rows.Next() {
		dir := new(svc.Directory)
		dir.Files = make(map[string]*svc.File, 0)
		dir.Dirs = make(map[string]*svc.Directory, 0)
		if err := rows.Scan(
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
			&dir.DriveID,
			&dir.Size,
			&dir.Path,
			&dir.ServerPath,
			&dir.ClientPath,
			&dir.BackupPath,
			&dir.Registered,
			&dir.Protected,
			&dir.AuthType,
			&dir.Key,
			&dir.Overwrite,
			&dir.LastSync,
			&dir.Endpoint,
			&dir.ParentID,
			&dir.Root,
			&dir.RootPath,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "no rows returned")
				continue
			}
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

func (q *Query) IsDirRegistered(dirID string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	var isRegistered bool
	if err := q.Conn.QueryRow(IsDirRegisteredQuery, dirID).Scan(&isRegistered); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return isRegistered, nil
}

// ------ drives --------------------------------

// get information about a user drive from the database
//
// drive returns nil if no information is available
func (q *Query) GetDrive(driveID string) (*svc.Drive, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	drv := new(svc.Drive)
	if err := q.Conn.QueryRow(FindDriveQuery, driveID).Scan(
		&drv.ID,
		&drv.OwnerName,
		&drv.OwnerID,
		&drv.TotalSize,
		&drv.UsedSpace,
		&drv.FreeSpace,
		&drv.Protected,
		&drv.Key,
		&drv.AuthType,
		&drv.IsLoaded,
		&drv.RootPath,
		&drv.RootID,
		&drv.Registered,
		&drv.RecycleBin,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, "no rows returned")
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return drv, nil
}

// return all drives from the database
func (q *Query) GetDrives() ([]*svc.Drive, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllDrivesQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %v", err)
	}
	defer rows.Close()

	drives := make([]*svc.Drive, 0)
	for rows.Next() {
		drv := new(svc.Drive)
		drv.Root = new(svc.Directory)
		drv.Root.Files = make(map[string]*svc.File, 0)
		drv.Root.Dirs = make(map[string]*svc.Directory, 0)
		drv.SyncIndex = new(svc.SyncIndex)
		if err := rows.Scan(
			&drv.ID,
			&drv.OwnerName,
			&drv.OwnerID,
			&drv.TotalSize,
			&drv.UsedSpace,
			&drv.FreeSpace,
			&drv.Protected,
			&drv.Key,
			&drv.AuthType,
			&drv.IsLoaded,
			&drv.RootPath,
			&drv.RootID,
			&drv.Registered,
			&drv.RecycleBin,
		); err != nil {
			if err == sql.ErrNoRows {
				q.log.Log(logger.INFO, "no rows returned")
				return nil, nil
			}
			return nil, fmt.Errorf("unable to query for drive: %v", err)
		}
		drives = append(drives, drv)
	}
	return drives, nil
}

// find a drive using the given userID.
// drive will be nil if not found.
func (q *Query) GetDriveByUserID(userID string) (*svc.Drive, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	drv := new(svc.Drive)
	if err := q.Conn.QueryRow(FindDriveByUserID, userID).Scan(
		&drv.ID,
		&drv.OwnerName,
		&drv.OwnerID,
		&drv.TotalSize,
		&drv.UsedSpace,
		&drv.FreeSpace,
		&drv.Protected,
		&drv.Key,
		&drv.AuthType,
		&drv.IsLoaded,
		&drv.RootPath,
		&drv.RootID,
		&drv.Registered,
		&drv.RecycleBin,
	); err != nil {
		if err == sql.ErrNoRows {
			q.log.Log(logger.INFO, "no rows returned")
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return drv, nil
}

// get the drive ID for a given userID
func (q *Query) GetDriveIDFromUserID(userID string) (string, error) {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	var id string
	err := q.Conn.QueryRow(FindUsersDriveIDQuery, userID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query: %v", err)
	} else if err == sql.ErrNoRows {
		q.log.Log(logger.INFO, fmt.Sprintf("no drive associated with user %v", userID))
		return "", nil
	}
	return id, nil
}
