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
		db.Close()
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
			createData DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id INT AUTO_INCREMENT PRIMARY KEY,
			owner VARCHAR(255) NOT NULL,
			createData DATETIME NOT NULL,
			opened BOOLEAN NOT NULL,
			FOREIGN KEY (owner) REFERENCES users(openid)
		);`,
		`CREATE TABLE IF NOT EXISTS scores (
			id INT AUTO_INCREMENT PRIMARY KEY,
			openid VARCHAR(255) NOT NULL,
			roomId INT NOT NULL,
			score INT NOT NULL,
			createData DATETIME NOT NULL,
			FOREIGN KEY (openid) REFERENCES users(openid),
			FOREIGN KEY (roomId) REFERENCES rooms(id)
		);`,
		`CREATE TABLE IF NOT EXISTS records (
			id INT AUTO_INCREMENT PRIMARY KEY,
			roomId INT NOT NULL,
			score INT NOT NULL,
			fromUser VARCHAR(255) NOT NULL,
			toUser VARCHAR(255) NOT NULL,
			createData DATETIME NOT NULL,
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

// 注册用户
func RegisterUser(openid string, nickname string) error {
	_, err := db.Exec("INSERT INTO users (openid, nickname, createData) VALUES (?, ?, NOW())", openid, nickname)
	if err != nil {
		fmt.Println("Error registering user:", err)
		return err
	}
	return nil
}

// 查询用户房间
func QueryUserRoom(openid string) (int, error) {
	var roomId int
	err := db.QueryRow("SELECT roomId FROM users WHERE openid = ?", openid).Scan(&roomId)
	if err != nil {
		fmt.Println("Error querying user room:", err)
		return 0, err
	}
	return roomId, nil
}

// 查询历史战绩
func QueryHistory(openid string, page int, pageSize int) ([]int, error) {
	var scores []int
	rows, err := db.Query("SELECT score FROM scores WHERE openid =? ORDER BY createData DESC LIMIT ?,?", openid, (page-1)*pageSize, pageSize)
	if err != nil {
		fmt.Println("Error querying history:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var score int
		err = rows.Scan(&score)
		if err != nil {
			fmt.Println("Error scanning history:", err)
			return nil, err
		}
	}
	return scores, nil
}

// 创建房间
func CreateRoom(openid string) (int, error) {
	var roomId int
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.QueryRow("INSERT INTO rooms (owner, createData, opened) VALUES (?, NOW(), 1) RETURNING id", openid).Scan(&roomId)
	if err != nil {
		fmt.Println("Error creating room:", err)
		return 0, err
	}
	_, err = tx.Exec("UPDATE users SET roomId = ? WHERE openid = ?", roomId, openid)
	if err != nil {
		fmt.Println("Error updating user room:", err)
		return 0, err
	}
	return roomId, nil
}

// 获取房间用户列表

// 获取房间分数列表

// 退出房间

// 解散房间(房主)

// 修改昵称

// 计分

// 加入房间
