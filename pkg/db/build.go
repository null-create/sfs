package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(dbName string, pathToNewDB string) error {
	switch dbName {
	case "users":
		NewTable(pathToNewDB, CreateUserTable)
	case "drives":
		NewTable(pathToNewDB, CreateDriveTable)
	case "directories":
		NewTable(pathToNewDB, CreateDirectoryTable)
	case "files":
		NewTable(pathToNewDB, CreateFileTable)
	default:
		return fmt.Errorf("unsupported database: %v", dbName)
	}
	return nil
}

// create a new table with the given query
func NewTable(path string, query string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("[ERROR] failed to create table: \n%v\n", err)
	}
}

// initialize server databases
func InitDBs(dbPath string) error {
	// make sure there's no databases where we want to create in
	entries, err := os.ReadDir(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read service database directory: %v", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("service database directory not empty! %v", entries)
	}

	dbs := []string{"files", "directories", "users", "drives"}
	for _, dbName := range dbs {
		if err := NewDB(dbName, filepath.Join(dbPath, dbName)); err != nil {
			return err
		}
	}
	return nil
}

// initialize client databases
func InitClientDBs(dbPath string) error {
	entries, err := os.ReadDir(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read service database directory: %v", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("service database directory not empty! %v", entries)
	}

	dbs := []string{"users", "files", "directories"}
	for _, dbName := range dbs {
		if err := NewDB(dbName, filepath.Join(dbPath, dbName)); err != nil {
			return err
		}
	}
	return nil
}
