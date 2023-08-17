package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	CreateUserTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Users')
	BEGIN
			CREATE TABLE Users (
					id VARCHAR(50) PRIMARY KEY,
					name VARCHAR(255),
					username VARCHAR(50),
					email VARCHAR(255),
					password VARCHAR(100),
					last_login_date DATETIME,
					is_admin BIT,
					total_files INT,
					total_directories INT
			);
	END`

	CreateDriveTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Storage')
	BEGIN
			CREATE TABLE Drives (
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
	END
	`

	CreateDirectoryTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Storage')
	BEGIN
			CREATE TABLE Directories (
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
	END
	`

	CreateFileTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Files')
	BEGIN
			CREATE TABLE Files (
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
	END
	`
)

func NewDB(dbName string, pathToNewDB string) {
	if dbName == "files" {
		new(pathToNewDB, CreateFileTable)
	} else if dbName == "directories" {
		new(pathToNewDB, CreateDirectoryTable)
	} else if dbName == "users" {
		new(pathToNewDB, CreateUserTable)
	} else if dbName == "drives" {
		new(pathToNewDB, CreateDirectoryTable)
	} else {
		log.Printf("[DEBUG] unknown database category: %s", dbName)
	}
}

func new(path string, query string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()
	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("[ERROR] failed to create file table: \n%v\n", err)
	}
}
