package db

import (
	"database/sql"
	"fmt"
	"scoringMP/config"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("mysql", config.Config.Mysql)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return err
	}
	return nil
}

func CreateTables() error {
	sqlStatements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			openid VARCHAR(255) PRIMARY KEY,
			nickname VARCHAR(255) NOT NULL,
			roomId INT,
			avatarUrl VARCHAR(255) NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id INT AUTO_INCREMENT PRIMARY KEY
		);`,
		`CREATE TABLE IF NOT EXISTS scores (
			id INT AUTO_INCREMENT PRIMARY KEY,
			openid VARCHAR(255) NOT NULL,
			roomId INT NOT NULL,
			score INT NOT NULL,
			FOREIGN KEY (openid) REFERENCES users(openid),
			FOREIGN KEY (roomId) REFERENCES rooms(id)
		);`,
		`CREATE TABLE IF NOT EXISTS records (
			id INT AUTO_INCREMENT PRIMARY KEY,
			roomId INT NOT NULL,
			score INT NOT NULL,
			fromUser VARCHAR(255) NOT NULL,
			toUser VARCHAR(255) NOT NULL,
			FOREIGN KEY (roomId) REFERENCES rooms(id),
			FOREIGN KEY (fromUser) REFERENCES users(openid),
			FOREIGN KEY (toUser) REFERENCES users(openid)
		);`,
	}
	for _, stmt := range sqlStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			fmt.Println("Error creating tables")
			return err
		}
	}
	return nil
}
