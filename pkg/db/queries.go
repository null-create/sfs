package db

// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)

const (

	// --------general table creation --------------------------------------

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
			checksum VARCHAR(255),
			algorithm VARCHAR(50)
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
			root_path VARCHAR(255)
		);
	`

	CreateDriveTable string = `
		CREATE TABLE IF NOT EXISTS Drives (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner VARCHAR(50),
			total_space DECIMAL(18, 2),
			used_space DECIMAL(18, 2),
			free_space DECIMAL(18, 2),
			protected BIT,
			key VARCHAR(100),
			auth_type VARCHAR(50),
			drive_root VARCHAR(255)
		);
`
	CreateUserTable string = `
		CREATE TABLE IF NOT EXISTS Users (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			username VARCHAR(50),
			email VARCHAR(255),
			password VARCHAR(100),
			last_login DATETIME,
			is_admin BIT,
			total_files INT,
			total_directories INT
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
			checksum, 
			algorithm
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

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
			owner,
			total_space,
			used_space,
			free_space,
			protected,
			key,
			auth_type,
			drive_root
		);
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	AddUserQuery string = `
		INSERT OR IGNORE INTO Users (
			id, 
			name, 
			username, 
			email, 
			password, 
			last_login, 
			is_admin, 
			total_files, 
			total_directories
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

	// ------- update file, user, directory, and drive entries -------

	// remove a user or file iff they (or the file) already exists in the database
	RemoveUserQuery string = `
		DELETE FROM Users WHERE id = '?' 
		AND EXISTS (SELECT 1 FROM Users WHERE id = '?');`

	RemoveFileQuery string = `
		DELETE FROM Files WHERE id = '?' 
		AND EXISTS (SELECT 1 FROM Users WHERE id = '?');`

	// ---------- SELECT statements for searching -------------------------------

	// NOTE: no limits are set on these queries because the id's are UUIDs, so we assume
	// there will only be one entry with this id in the database

	FindFileQuery   string = `SELECT * FROM Files  WHERE id = '?';`
	FindDirQuery    string = `SELECT * FROM Directories WHERE id = '?'`
	FindDriverQuery string = `SELECT * FROM Drives WHERE id = '?'`
	FindUserQuery   string = `SELECT * FROM Users WHERE id = '?'`

	// ---------- SELECT statements for confirming existance -------------------

	UserExistsQuery string = `SELECT EXISTS(SELECT 1 FROM Users WHERE id = ?)`
)
