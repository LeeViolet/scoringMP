package db

import (
	"database/sql"
	"errors"
	"fmt"
	"scoringMP/config"
	"scoringMP/model"

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

// 查询用户
func QueryUser(openid string) (model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT * FROM users WHERE openid =?", openid).Scan(&user.Openid, &user.Nickname, &user.RoomId, &user.CreateData)
	return user, err
}

// 注册用户
func RegisterUser(openid string, nickname string) error {
	_, err := db.Exec("INSERT INTO users (openid, nickname, createData) VALUES (?, ?, NOW())", openid, nickname)
	return err
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
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE openid =?", openid).Scan(&count)
	if err != nil {
		fmt.Println("Error querying user count:", err)
		return 0, err
	}
	if count == 0 {
		fmt.Println("用户不存在")
		return 0, errors.New("用户不存在")
	}
	result, err := tx.Exec("INSERT INTO rooms (owner, createData, opened) VALUES (?, NOW(), 1)", openid)
	if err != nil {
		fmt.Println("Error creating room:", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting last insert ID:", err)
		return 0, err
	}
	_, err = tx.Exec("UPDATE users SET roomId = ? WHERE openid = ?", id, openid)
	if err != nil {
		fmt.Println("Error updating user room:", err)
		return 0, err
	}
	return int(id), nil
}

// 获取房间用户列表
func GetRoomUsers(roomId int) ([]model.User, error) {
	var users []model.User
	// 检查房间是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rooms WHERE id =?", roomId).Scan(&count)
	if err != nil {
		fmt.Println("Error querying room count:", err)
		return nil, err
	}
	if count == 0 {
		fmt.Println("房间不存在")
		return nil, errors.New("房间不存在")
	}
	rows, err := db.Query("SELECT * FROM users WHERE roomId =?", roomId)
	if err != nil {
		fmt.Println("Error querying room users:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.Openid, &user.Nickname, &user.RoomId, &user.CreateData)
		if err != nil {
			fmt.Println("Error scanning room users:", err)
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// 获取房间分数列表
func GetRoomRecord(roomId int) ([]model.Record, error) {
	var records []model.Record
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rooms WHERE id =?", roomId).Scan(&count)
	if err != nil {
		fmt.Println("Error querying room count:", err)
		return nil, err
	}
	if count == 0 {
		fmt.Println("房间不存在")
		return nil, errors.New("房间不存在")
	}
	rows, err := db.Query("SELECT * FROM records WHERE roomId =?", roomId)
	if err != nil {
		fmt.Println("Error querying room scores:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record model.Record
		err = rows.Scan(&record.Id, &record.RoomId, &record.Score, &record.FromUser, &record.ToUser, &record.CreateData)
		if err != nil {
			fmt.Println("Error scanning room scores:", err)
			return nil, err
		}
	}
	return records, nil
}

// 退出房间
func QuitRoom(openid string) error {
	var roomId sql.NullInt64
	err := db.QueryRow("SELECT roomId FROM users WHERE openid =?", openid).Scan(&roomId)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("用户不存在")
			return errors.New("用户不存在")
		}
		fmt.Println("Error querying user room:", err)
		return err
	}
	if roomId.Valid {
		var owner string
		err = db.QueryRow("SELECT owner FROM rooms WHERE id =?", roomId.Int64).Scan(&owner)
		if err != nil {
			fmt.Println("Error querying room owner:", err)
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			fmt.Println("Error starting transaction:", err)
			return err
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
		// 如果是房主退出，将所有人的 roomId 设置为 null
		if owner == openid {
			_, err = tx.Exec("UPDATE users SET roomId = NULL WHERE roomId =?", roomId.Int64)
			if err != nil {
				fmt.Println("Error updating user room:", err)
				return err
			}
			_, err = tx.Exec("UPDATE rooms SET opened = 0 WHERE id =?", roomId.Int64)
			if err != nil {
				fmt.Println("Error updating room opened:", err)
				return err
			}
		} else {
			_, err = tx.Exec("UPDATE users SET roomId = NULL WHERE openid =?", openid)
		}
	}
	return nil
}

// 修改昵称
func ModifyNickname(openid string, nickname string) error {
	_, err := db.Exec("UPDATE users SET nickname =? WHERE openid =?", nickname, openid)
	if err != nil {
		fmt.Println("Error modifying nickname:", err)
		return err
	}
	return nil
}

// 计分
func AddRecords(roomId int, fromUser string, toUser string, score int) error {
	_, err := db.Exec("INSERT INTO Records (roomId, score, fromUser, toUser, createData) VALUES (?, ?, ?, ?, NOW())", roomId, score, fromUser, toUser)
	if err != nil {
		fmt.Println("Error inserting record:", err)
		return err
	}
	return nil
}

// 加入房间
