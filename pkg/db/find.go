package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// Query to check user existence
func (q *Query) UserExists(userID string) (bool, error) {
	q.WhichDB("users")

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

// get user data from database
func (q *Query) GetUser(userID string) (*auth.User, error) {
	q.WhichDB("users")
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
			log.Printf("[DEBUG] no rows returned: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] unable to execute query: %v", err)
	}

	return user, nil
}

// get a userID from a driveID.
// will return an empty string if no userID is found with this driveID
func (q *Query) GetUserIDFromDriveID(driveID string) (string, error) {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	if q.Debug {
		log.Printf("[DEBUG] querying for userID using driveID %s", driveID)
	}

	var userID string
	if err := q.Conn.QueryRow(FindUsersDriveIDQuery, driveID).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] no user associated with driveID %s", driveID)
			return "", nil
		} else {
			return "", fmt.Errorf("failed to query for userID: %v", err)
		}
	}
	return userID, nil
}

// populates a slice of *auth.User structs of all available users from the database
func (q *Query) GetUsers() ([]*auth.User, error) {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
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
				log.Printf("[DEBUG] no rows returned: %v", err)
				return nil, nil
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

// ----- files ----------------------------------

// retrieve file metadata from the database
//
// file returns nil if no result is available
func (q *Query) GetFile(fileID string) (*svc.File, error) {
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
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.Endpoint,
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

// find a file in the database by searching with its path
func (q *Query) GetFileByPath(filePath string) (*svc.File, error) {
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
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned (path=%s): %v", filePath, err)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// get a file by name. returns nil if no file is found in the db.
func (q *Query) GetFileByName(fileName string) (*svc.File, error) {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	file := new(svc.File)
	if err := q.Conn.QueryRow(FindFileByNameQuery, fileName).Scan(
		&file.ID,
		&file.Name,
		&file.OwnerID,
		&file.DirID,
		&file.DriveID,
		&file.Protected,
		&file.Key,
		&file.LastSync,
		&file.Path,
		&file.ServerPath,
		&file.ClientPath,
		&file.Endpoint,
		&file.CheckSum,
		&file.Algorithm,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned (file name=%s): %v", fileName, err)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}
	return file, nil
}

// retrieves a file ID using a given file path
func (q *Query) GetFileID(filePath string) (string, error) {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	var fileID string
	if err := q.Conn.QueryRow(FindFileIDWithPathQuery, filePath).Scan(&fileID); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned (path=%s): %v", filePath, err)
			return "", nil
		}
		return "", fmt.Errorf("failed to get file ID: %v", err)
	}
	return fileID, nil
}

// populate a slice of *svc.File structs from *all*
// files in the files database
func (q *Query) GetFiles() ([]*svc.File, error) {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllFilesQuery)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
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
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.Endpoint,
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

// populate a slice of *svc.File structs for all files
// associated with the given user. will return an empty slice
// if no files are found.
func (q *Query) GetUsersFiles(userID string) ([]*svc.File, error) {
	q.WhichDB("files")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllUsersFilesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
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
			&file.Protected,
			&file.Key,
			&file.LastSync,
			&file.Path,
			&file.ServerPath,
			&file.ClientPath,
			&file.Endpoint,
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

// ----------- directories --------------------------------

// retrieve information about a users directory from the database
//
// returns nil if no information is available
func (q *Query) GetDirectory(dirID string) (*svc.Directory, error) {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	d := new(svc.Directory)
	d.Files = make(map[string]*svc.File, 0)
	d.Dirs = make(map[string]*svc.Directory, 0)

	if err := q.Conn.QueryRow(FindDirQuery, dirID).Scan(
		&d.ID,
		&d.Name,
		&d.OwnerID,
		&d.DriveID,
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

// find a directory by name. returns nil if no directory is found.
func (q *Query) GetDirectoryByName(dirName string) (*svc.Directory, error) {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	d := new(svc.Directory)
	d.Files = make(map[string]*svc.File, 0)
	d.Dirs = make(map[string]*svc.Directory, 0)

	if err := q.Conn.QueryRow(FindDirByNameQuery, dirName).Scan(
		&d.ID,
		&d.Name,
		&d.OwnerID,
		&d.DriveID,
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
			log.Printf("[DEBUG] no rows found for dir: %s", dirName)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] query failed: %v", err)
	}
	return d, nil
}

// geta a slice of *all* directories on the server, regardless
// of owner. only for admin users.
func (q *Query) GetDirectories(limit int) ([]*svc.Directory, error) {
	q.WhichDB("directories")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllDirsQuery)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	dirs := make([]*svc.Directory, 0)
	for rows.Next() {
		dir := new(svc.Directory)
		dir.Files = make(map[string]*svc.File, 0)
		dir.Dirs = make(map[string]*svc.Directory, 0)
		if err := q.Conn.QueryRow(FindAllQuery, "Directories").Scan(
			&dir.ID,
			&dir.Name,
			&dir.OwnerID,
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

// ------ drives --------------------------------

// get information about a user drive from the database
//
// drive returns nil if no information is available
func (q *Query) GetDrive(driveID string) (*svc.Drive, error) {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	d := new(svc.Drive)
	if err := q.Conn.QueryRow(FindDriveQuery, driveID).Scan(
		&d.ID,
		&d.OwnerName,
		&d.OwnerID,
		&d.TotalSize,
		&d.UsedSpace,
		&d.FreeSpace,
		&d.Protected,
		&d.Key,
		&d.AuthType,
		&d.DriveRoot,
		&d.RootID,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] query failed: %v", err)
	}

	return d, nil
}

// return all drives from the database
func (q *Query) GetDrives() ([]*svc.Drive, error) {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	rows, err := q.Conn.Query(FindAllDrivesQuery)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] unable to query: %v", err)
	}
	defer rows.Close()

	drives := make([]*svc.Drive, 0)
	for rows.Next() {
		drive := new(svc.Drive)
		drive.Root = new(svc.Directory)
		drive.SyncIndex = new(svc.SyncIndex)
		if err := rows.Scan(
			&drive.ID,
			&drive.OwnerName,
			&drive.OwnerID,
			&drive.TotalSize,
			&drive.UsedSpace,
			&drive.FreeSpace,
			&drive.Protected,
			&drive.Key,
			&drive.AuthType,
			&drive.DriveRoot,
			&drive.RootID,
		); err != nil {
			return nil, fmt.Errorf("[ERROR] unable to query for drive: %v", err)
		}
		drives = append(drives, drive)
	}
	if len(drives) == 0 {
		log.Print("[INFO] no drives found")
	}
	return drives, nil
}

// find a drive using the given userID.
// drive will be nil if not found.
func (q *Query) GetDriveByUserID(userID string) (*svc.Drive, error) {
	q.WhichDB("drives")
	q.Connect()
	defer q.Close()

	d := new(svc.Drive)
	if err := q.Conn.QueryRow(FindDriveByUserID, userID).Scan(
		&d.ID,
		&d.OwnerName,
		&d.OwnerID,
		&d.TotalSize,
		&d.UsedSpace,
		&d.FreeSpace,
		&d.Protected,
		&d.Key,
		&d.AuthType,
		&d.DriveRoot,
		&d.RootID,
	); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[DEBUG] no rows returned: %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("[ERROR] query failed: %v", err)
	}
	return d, nil
}

// get the drive ID for a given userID
func (q *Query) GetDriveID(userID string) (string, error) {
	q.WhichDB("users")
	q.Connect()
	defer q.Close()

	var id string
	err := q.Conn.QueryRow(FindUsersDriveIDQuery, userID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query: %v", err)
	} else if err == sql.ErrNoRows {
		log.Printf("no drive associated with user %v", userID)
		return "", nil
	}
	return id, nil
}
