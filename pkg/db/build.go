package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func MakeNewDatabase(pathToNewDb string, dbName string) {
	db, err := sql.Open("sqlite3", pathToNewDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a table in the database
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, user_name TEXT NOT NULL)", dbName)
	if err != nil {
		log.Fatal(err)
	}
	result, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SQLite3 database created successfully! \n%v\n", result)
}
