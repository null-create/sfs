package db

// Insert query with ON CONFLICT IGNORE (SQLite equivalent of upsert)

const (

	// --------general table creation --------------------------------------
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

	CreateDirectoryTable string = `
		CREATE TABLE IF NOT EXISTS Directories (
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

	CreateFileTable string = `
		CREATE TABLE IF NOT EXISTS Files (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255),
			owner VARCHAR(50),
			protected BIT,
			key VARCHAR(100),
			path VARCHAR(255),
			last_sync DATETIME,
			server_path VARCHAR(255),
			client_path VARCHAR(255),
			checksum VARCHAR(255),
			algorithm VARCHAR(50)
		);
	`

	DropTableQuery string = `DROP TABLE IF EXISTS ?;`

	// ------- file, user, directory, and drive additions ----------------
	AddFileQuery string = `
	INSERT OR IGNORE INTO Files (
		id, 
		name, 
		owner, 
		protected, 
		key, 
		path, 
		server_path, 
		client_path, 
		checksum, 
		algorithm
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

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

	AddDirQuery string = ``

	AddDriverQuery string = ``

	// ------- update file, user, directory, and drive entries -------

	// remove a user or file iff they (or the file) already exists in the database
	RemoveUserQuery string = `
		DELETE FROM Users WHERE id = '?' 
		AND EXISTS (SELECT 1 FROM Users WHERE id = '?');`

	RemoveFileQuery string = `
		DELETE FROM Files WHERE id = '?' 
		AND EXISTS (SELECT 1 FROM Users WHERE id = '?');`

	// ---------- SELECT statements for searching -------------------------------

)
