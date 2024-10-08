package db

// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)

const (

	// --------table creation --------------------------------------

	CreateFileTable string = `
		CREATE TABLE IF NOT EXISTS Files (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner_id VARCHAR(50),
			directory_id VARCHAR(50),
			drive_id VARCHAR(50),
			mode VARCHAR(50),
			size INTEGER,
			server_backup BIT,
			local_backup BIT,
			protected BIT,
			key VARCHAR(100),
			last_sync DATETIME,
			path VARCHAR(255),
			server_path VARCHAR(255),
			client_path VARCHAR(255),
			backup_path VARCHAR(255),
			registered BIT,
			endpoint VARCHAR(255),
			checksum VARCHAR(255),
			algorithm VARCHAR(50),
			UNIQUE(id)
		);`

	CreateDirectoryTable string = `
		CREATE TABLE IF NOT EXISTS Directories (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner_id VARCHAR(50),
			drive_id VARCHAR(50),
			size DECIMAL(18, 2),
			path VARCHAR(255),
			server_path VARCHAR(255),
			client_path VARCHAR(255),
			backup_path VARCHAR(255),
			registered BIT,
			protected BIT,
			auth_type VARCHAR(50),
			key VARCHAR(100),
			overwrite BIT,
			last_sync DATETIME,
			endpoint VARCHAR(255),
			parent_id VARCHAR(50),
			drive_root VARCHAR(255),
			root_path VARCHAR(255),
			UNIQUE(id)
		);
	`

	CreateDriveTable string = `
		CREATE TABLE IF NOT EXISTS Drives (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner_id VARCHAR(50),
			total_space DECIMAL(18, 2),
			used_space DECIMAL(18, 2),
			free_space DECIMAL(18, 2),
			protected BIT,
			key VARCHAR(100),
			auth_type VARCHAR(50),
			is_loaded BIT,
			root_path VARCHAR(255),
			root_id VARCHAR(50),
			registered BIT, 
			recycle_bin VARCHAR(255),
			UNIQUE(id)
		);`

	CreateUserTable string = `
		CREATE TABLE IF NOT EXISTS Users (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			username VARCHAR(50),
			email VARCHAR(255),
			password VARCHAR(100),
			last_login DATETIME,
			is_admin BIT,
			sf_path VARCHAR(255),
			drive_id VARCHAR(255),  
			total_files INT,
			total_directories INT,
			root VARCHAR(255),
			UNIQUE(id)
		);`

	// ------- file, user, directory, and drive additions ----------------

	AddFileQuery string = `
		INSERT OR IGNORE INTO Files (
			id,
			name,
			owner_id,
			directory_id,
			drive_id,
			mode, 
			size, 
			server_backup, 
			local_backup,
			protected,
			key,
			last_sync,
			path,
			server_path,
			client_path,
			backup_path,
			registered,
			endpoint,
			checksum,
			algorithm
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	AddDirQuery string = `
		INSERT OR IGNORE INTO Directories (
			id,
			name,
			owner_id,
			drive_id,
			size,
			path,
			server_path,
			client_path,
			backup_path,
			registered,
			protected,
			auth_type,
			key,
			overwrite,
			last_sync,
			endpoint,
			parent_id, 
			drive_root, 
			root_path
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	AddDriveQuery string = `
		INSERT OR IGNORE INTO Drives (
			id,
			name,
			owner_id,
			total_space,
			used_space,
			free_space,
			protected,
			key,
			auth_type,
			is_loaded,
			root_path,
			root_id,
			registered, 
			recycle_bin
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	AddUserQuery string = `
		INSERT OR IGNORE INTO Users (
			id, 
			name, 
			username, 
			email, 
			password, 
			last_login, 
			is_admin,
			sf_path,
			drive_id, 
			total_files, 
			total_directories,
			root
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// ------- update file, user, directory, and drive entries -------

	UpdateFileQuery string = `
		UPDATE Files
		SET id = ?, 
				name = ?, 
				owner_id = ?,
				directory_id = ?,
				drive_id = ?, 
				mode = ?, 
				size = ?, 
				server_backup = ?, 
				protected = ?, 
				key = ?,
				last_sync = ?, 
				path = ?, 
				server_path = ?, 
				client_path = ?,
				backup_path = ?,
				registered = ?,  
				endpoint = ?,  
				checksum = ?, 
				algorithm = ?
		WHERE id = ?;`

	UpdateDirQuery string = `
		UPDATE Directories
		SET id = ?,
				name = ?,
				owner_id = ?,
				drive_id = ?,
				size = ?,
				path = ?,
				server_path = ?,
				client_path = ?,
				backup_path = ?,
				registered = ?,
				protected = ?,
				auth_type = ?,
				key = ?,
				overwrite = ?,
				last_sync = ?, 
				endpoint = ?,
				parent_id = ?,
				drive_root = ?, 
				root_path = ?
		WHERE id = ?;`

	UpdateDriveQuery string = `
		UPDATE Drives
		SET id = ?,
				name = ?,
				owner_id = ?,
				total_space = ?,
				used_space = ?,
				free_space = ?,
				protected = ?,
				key = ?,
				auth_type = ?,
				is_loaded = ?,
				root_path = ?,
				root_id = ?,
				registered = ?,
				recycle_bin = ?
		WHERE id = ?;`

	UpdateUserQuery string = `
		UPDATE Users
		SET id = ?, 
				name = ?, 
				username = ?, 
				email = ?, 
				password = ?, 
				last_login = ?,
				is_admin = ?,
				sf_path = ?,
				drive_id = ?,
				total_files = ?,
				total_directories = ?,
				root = ?
		WHERE id = ?;`

	// ----------- Removal queries remove the row iff they exist

	RemoveFileQuery string = `
		DELETE FROM Files WHERE id = ? 
		AND EXISTS (SELECT 1 FROM Files WHERE id = ?);`

	RemoveDirectoryQuery string = `
		DELETE FROM Directories WHERE id = ? 
		AND EXISTS (SELECT 1 FROM Directories WHERE id = ?);`

	RemoveDriveQuery string = `
		DELETE FROM Drives WHERE id = ? 
		AND EXISTS (SELECT 1 FROM Drives WHERE id = ?);`

	RemoveUserQuery string = `
		DELETE FROM Users WHERE id = ? 
		AND EXISTS (SELECT 1 FROM Users WHERE id=?);`

	DropUserTableQuery string = `DROP TABLE IF EXISTS Users;`

	DropDrivesTableQuery string = `DROP TABLE IF EXISTS Drives;`

	DropDirectoriesTableQuery string = `DROP TABLE IF EXISTS Directories;`

	DropFilesTableQuery string = `DROP TABLE IF EXISTS Files;`

	// ---------- SELECT statements for searching -------------------------------

	// general
	FindAllQuery       string = `SELECT * FROM ?;`
	FindWithLimitQuery string = `SELECT * FROM ? LIMIT ?;`
	FindQuery          string = `SELECT * FROM ? WHERE id = ?;`

	// find all
	FindAllUsersQuery  string = `SELECT * FROM Users;`
	FindAllDrivesQuery string = `SELECT * FROM Drives;`
	FindAllDirsQuery   string = `SELECT * FROM Directories;`
	FindAllFilesQuery  string = `SELECT * FROM Files;`

	// find specific
	FindDirByNameQuery           string = `SELECT * FROM Directories WHERE name = ?;`
	FindAllUsersFilesQuery       string = `SELECT * FROM Files WHERE owner_id = ?;`
	FindFileIDWithPathQuery      string = `SELECT id FROM Files WHERE path = ?;`
	FindFileQuery                string = `SELECT * FROM Files WHERE id = ?;`
	FindFilesByNameQuery         string = `SELECT * FROM Files WHERE name = ?;`
	FindFileByPathQuery          string = `SELECT * FROM Files WHERE path = ?;`
	FindFilesByDirIDQuery        string = `SELECT * FROM Files WHERE directory_id =?;`
	FindFilesByDriveIDQuery      string = `SELECT * FROM Files WHERE drive_id = ?;`
	FindAllBackedUpFilesQuery    string = `SELECT * FROM Files WHERE backup = 1;`
	FindDirQuery                 string = `SELECT * FROM Directories WHERE id = ?;`
	FindAllUsersDirectoriesQuery string = `SELECT * FROM Directories WHERE owner_id = ?;`
	FindDirsByNameQuery          string = `SELECT * FROM Directories WHERE name = ?`
	FindDirsByDriveIDQuery       string = `SELECT * FROM Directories WHERE drive_id = ?;`
	FindDirByPathQuery           string = `SELECT * FROM Directories WHERE path = ?;`
	FindDirsByParentIDQuery      string = `SELECT * FROM Directories WHERE parent_id = ?;`
	FindDirIDByPathQuery         string = `SELECT id FROM Directories WHERE path = ?;`
	FindDriveQuery               string = `SELECT * FROM Drives WHERE id = ?;`
	FindDriveByUserID            string = `SELECT * FROM Drives WHERE owner_id = ?;`
	FindUserQuery                string = `SELECT * FROM Users WHERE id = ?;`
	FindUsersDriveIDQuery        string = `SELECT drive_id FROM Users WHERE id = ?;`
	FindUsersIDWithDriveIDQuery  string = `SELECT owner_id FROM Drives WHERE id = ?;`

	// find by date ranges
	FindFilesAfterQuery string = `SELECT * FROM Files WHERE last_sync > ?;`
	FindDirsAfterQuery  string = `SELECT * FROM Directories WHERE last_sync > ?;`

	// ---------- SELECT statements for confirming existance and registration  -------------------

	ExistsQuery           string = `SELECT EXISTS (SELECT 1 FROM ? WHERE id = '?');`
	IsFileRegisteredQuery string = `SELECT registered FROM Files WHERE id = ?;`
	IsDirRegisteredQuery  string = `SELECT registered FROM Directories WHERE id = ?;`
)
