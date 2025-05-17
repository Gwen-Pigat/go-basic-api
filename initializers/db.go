package initializers

import (
	"database/sql"
	"fmt"
	"os"
)

var DB *sql.DB

func ConnectDB(dbURI string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	err = SetupDB(db)
	if err != nil {
		return nil, err
	}
	fmt.Println(db)
	return db, nil
}

func SetupDB(db *sql.DB) error {
	_, err := db.Exec(`
		SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
		SET time_zone = "+00:00";
		CREATE DATABASE IF NOT EXISTS ` + os.Getenv("DB_NAME") + `
	`)
	if err != nil {
		return err
	}
	defer db.Close()
	DB, err = sql.Open("mysql", os.Getenv("DB_URI"))
	if err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}
	dbInit, err := os.ReadFile("initializers/db.sql")
	if err != nil {
		return err
	}
	_, err = DB.Exec(string(dbInit))
	if err != nil {
		return err
	}
	return nil
}

func ExecFlushDB(db *sql.DB) error {
	dbQuery := "DROP TABLE task;DROP TABLE user"
	if _, err := db.Exec(dbQuery); err != nil {
		return err
	}
	if err := SetupDB(db); err != nil {
		return err
	}
	return nil
}
