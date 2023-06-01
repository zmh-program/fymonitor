package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func ConnectDatabase() *sql.DB {
	db, err := sql.Open("mysql", viper.GetString("database.dsn"))
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS monitor (
			id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(256) NOT NULL,
			responses JSON,
			stamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		panic(err)
	}

	return db
}
