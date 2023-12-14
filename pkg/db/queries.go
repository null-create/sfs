package db

// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)

const (

	// --------table creation --------------------------------------

	CreateFileTable string = `
		CREATE TABLE IF NOT EXISTS Files (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner VARCHAR(50),
			protected BIT,
			key VARCHAR(100),
			last_sync DATETIME,
			path VARCHAR(255),
			server_path VARCHAR(255),
			client_path VARCHAR(255),
			endpoint VARCHAR(255),
			checksum VARCHAR(255),
			algorithm VARCHAR(50),
			UNIQUE(id)
		);`

	CreateDirectoryTable string = `
		CREATE TABLE IF NOT EXISTS Directories (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner VARCHAR(50),
			size DECIMAL(18, 2),
			path VARCHAR(255),
			protected BIT,
			auth_type VARCHAR(50),
			key VARCHAR(100),
			overwrite BIT,
			last_sync DATETIME,
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
			drive_root VARCHAR(255),
			root_id VARCHAR(50),
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

	DropTableQuery string = `DROP TABLE IF EXISTS ?;`

	// ------- file, user, directory, and drive additions ----------------

	AddFileQuery string = `
		INSERT OR IGNORE INTO Files (
			id, 
			name, 
			owner, 
			protected, 
			key,
			last_sync, 
			path, 
			server_path, 
			client_path,
			endpoint, 
			checksum, 
			algorithm
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	AddDirQuery string = `
		INSERT OR IGNORE INTO Directories (
			id,
			name,
			owner,
			size,
			path,
			protected,
			auth_type,
			key,
			overwrite,
			last_sync, 
			drive_root, 
			root_path
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
			drive_root,
			root_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
				owner = ?, 
				protected = ?, 
				key = ?,
				last_sync = ?, 
				path = ?, 
				server_path = ?, 
				client_path = ?,
				endpoint = ?,  
				checksum = ?, 
				algorithm = ?
		WHERE id = ?;`

	UpdateDirQuery string = `
		UPDATE Directories
		SET id = ?,
				name = ?,
				owner = ?,
				size = ?,
				path = ?,
				protected = ?,
				auth_type = ?,
				key = ?,
				overwrite = ?,
				last_sync = ?, 
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
				drive_root = ?,
				root_id = ?
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

	// Removal queries remove the row iff they exist

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
	FindDirByNameQuery          string = `SELECT * FROM Directories WHERE name = ?;`
	FindAllUsersFilesQuery      string = `SELECT * FROM Files WHERE owner = ?;`
	FindFileIDWithPathQuery     string = `SELECT id FROM Files WHERE path = ?;`
	FindFileQuery               string = `SELECT * FROM Files WHERE id = ?;`
	FindFileByNameQuery         string = `SELECT * FROM Files WHERE name = ?;`
	FindFileByPathQuery         string = `SELECT * FROM Files WHERE path = ?;`
	FindDirQuery                string = `SELECT * FROM Directories WHERE id = ?;`
	FindDriveQuery              string = `SELECT * FROM Drives WHERE id = ?;`
	FindDriveByUserID           string = `SELECT * FROM Drives WHERE owner_id = ?;`
	FindUserQuery               string = `SELECT * FROM Users WHERE id = ?;`
	FindUsersDriveIDQuery       string = `SELECT drive_id FROM Users WHERE id = ?;`
	FindUsersIDWithDriveIDQuery string = `SELECT owner FROM Drives WHERE id = ?;`

	// ---------- SELECT statements for confirming existance -------------------

	ExistsQuery string = `SELECT EXISTS (SELECT 1 FROM ? WHERE id = '?');`
)
