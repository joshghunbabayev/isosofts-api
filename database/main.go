package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func GetDatabase() (db *sql.DB) {
	db, err := sql.Open("sqlite", "./database/main.db")

	if err != nil {
		panic(err)
	}

	return db
}
