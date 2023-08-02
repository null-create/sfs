package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const createUserTable = `
	CREATE TABLE IF NOT EXISTS Users (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  name VARCHAR(64) NOT NULL,
	  hash VARCHAR(255) NOT NULL,
	)
`

func Migrate(dbDriver *sql.DB) {
	statement, err := dbDriver.Prepare(createUserTable)
	if err == nil {
		_, creationError := statement.Exec()
		if creationError == nil {
			log.Println("Table created successfully")
		} else {
			log.Println(creationError.Error())
		}
	} else {
		log.Println(err.Error())
	}
}
