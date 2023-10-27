package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func setupDB() (*sql.DB, error) {
	dsn := "root:password@tcp(localhost:33060)/itunes"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
