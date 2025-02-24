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
		fmt.Println("Error opening mysql:", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Error pinging mysql:", err)
		return err
	}
	// 创建数据库
	_, err = db.Exec(`CREATE DATABASE IF NOT EXISTS scoring
			CHARACTER SET utf8mb4
			COLLATE utf8mb4_unicode_ci
		;`)
	if err != nil {
		fmt.Println("Error creating database:", err)
		return err
	}
	db.Close()
	// 连接到数据库
	db, err = sql.Open("mysql", config.Config.Mysql+"scoring")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Error pinging database:", err)
		return err
	}
	return nil
}

func CreateTables() error {
	sqlStatements := []string{
		`CREATE DATABASE IF NOT EXISTS scoring
			CHARACTER SET utf8mb4
			COLLATE utf8mb4_unicode_ci
		;`,
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
			fmt.Println("Error creating tables", err)
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
func QueryUserRoom(openid string) (model.Room, error) {
	var room model.Room
	err := db.QueryRow("SELECT * FROM rooms WHERE id =(SELECT roomId FROM users WHERE openid =? AND opened = 1)", openid).Scan(&room.Id, &room.Owner, &room.CreateData, &room.Opened)
	return room, err
}

// 查询历史战绩
func QueryHistory(openid string) ([]int, error) {
	var scores []int
	rows, err := db.Query("SELECT score FROM scores WHERE openid =? ORDER BY createData DESC", openid)
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

// 加入房间
func JoinRoom(openid string, roomId int) error {
	// 检查房间是否关闭
	var opened bool
	err := db.QueryRow("SELECT opened FROM rooms WHERE id =?", roomId).Scan(&opened)
	if err != nil {
		fmt.Println("Error querying room opened:", err)
		return err
	}
	if !opened {
		return errors.New("room is closed")
	}
	// 检查用户是否已经在房间中
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE openid =? AND roomId =?", openid, roomId).Scan(&count)
	if err != nil {
		fmt.Println("Error querying user in room:", err)
		return err
	}
	if count > 0 {
		return errors.New("user already in room")
	}
	// 加入房间
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
	_, err = tx.Exec("UPDATE users SET roomId =? WHERE openid =?", roomId, openid)
	if err != nil {
		fmt.Println("Error updating user room:", err)
		return err
	}
	_, err = tx.Exec("INSERT INTO scores (openid, roomId, score, createData) VALUES (?,?,?, NOW())", openid, roomId, 0)
	if err != nil {
		fmt.Println("Error inserting user score:", err)
		return err
	}
	return nil
}

// 创建/回到房间
func CreateRoom(openid string) (int, error) {
	// 检查用户是否已经在房间中
	user, err := QueryUser(openid)
	if err != nil {
		fmt.Println("Error querying user:", err)
		return 0, err
	}
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
	if user.RoomId.Valid {
		return int(user.RoomId.Int64), nil
	}
	// 没有房间则创建房间
	result, err := tx.Exec("INSERT INTO rooms (owner, createData, opened) VALUES (?, NOW(), 1)", openid)
	if err != nil {
		fmt.Println("Error inserting room:", err)
		return 0, err
	}
	roomId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting room id:", err)
		return 0, err
	}
	_, err = tx.Exec("UPDATE users SET roomId =? WHERE openid =?", roomId, openid)
	if err != nil {
		fmt.Println("Error updating user room:", err)
		return 0, err
	}
	_, err = tx.Exec("INSERT INTO scores (openid, roomId, score, createData) VALUES (?,?,?, NOW())", openid, roomId, 0)

	return int(roomId), nil
}

// 检查房间是否关闭
func CheckRoom(roomId int) (bool, error) {
	var opened bool
	err := db.QueryRow("SELECT opened FROM rooms WHERE id =?", roomId).Scan(&opened)
	if err != nil {
		fmt.Println("Error querying room opened:", err)
		return false, err
	}
	return opened, nil
}

type UserScore struct {
	Openid   string `json:"openid"`
	Score    int    `json:"score"`
	Nickname string `json:"nickname"`
}

// 获取房间用户列表及其 score
func GetRoomUsers(roomId int) ([]UserScore, error) {
	var users []UserScore
	rows, err := db.Query(`
		SELECT u.openid, u.nickname, s.score
		FROM scores s
		JOIN users u ON s.openid = u.openid
		WHERE s.roomId =?
		ORDER BY s.createData DESC
	`, roomId)
	if err != nil {
		fmt.Println("Error querying room users:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user UserScore
		err = rows.Scan(&user.Openid, &user.Nickname, &user.Score)
		if err != nil {
			fmt.Println("Error scanning room users:", err)
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

type UserRecord struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Score    int    `json:"score"`
}

// 获取房间分数列表
func GetRoomRecords(roomId int) ([]UserRecord, error) {
	var records []UserRecord
	rows, err := db.Query(`
		SELECT u1.nickname AS fromUser, u2.nickname AS toUser, r.score
		FROM records r
		JOIN users u1 ON r.fromUser = u1.openid
		JOIN users u2 ON r.toUser = u2.openid
		WHERE r.roomId = ?
		ORDER BY r.createData DESC
	`, roomId)
	if err != nil {
		fmt.Println("Error querying room records:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var record UserRecord
		err = rows.Scan(&record.FromUser, &record.ToUser, &record.Score)
		if err != nil {
			fmt.Println("Error scanning room records:", err)
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// 退出房间
func QuitRoom(openid string, roomId int) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return false, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// 检查用户是否在房间内
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE openid =? AND roomId =?", openid, roomId).Scan(&count)
	if err != nil {
		fmt.Println("Error querying user in room:", err)
		return false, err
	}
	if count == 0 {
		return false, errors.New("user not in room")
	}
	// 检查用户是否是房主
	var room model.Room
	err = tx.QueryRow("SELECT * FROM rooms WHERE id =? AND opened = 1", roomId).Scan(&room.Id, &room.Owner, &room.CreateData, &room.Opened)
	if err != nil {
		fmt.Println("Error querying room:", err)
		return false, err
	}
	if room.Owner != openid {
		// 不是房主，直接退出
		_, err = tx.Exec("UPDATE users SET roomId = NULL WHERE openid =? AND roomId =?", openid, roomId)
		if err != nil {
			fmt.Println("Error updating user room:", err)
			return false, err
		}
		return false, nil
	} else {
		// 是房主，所有人退出房间，关闭房间
		_, err = tx.Exec("UPDATE users SET roomId = NULL WHERE roomId =?", roomId)
		if err != nil {
			fmt.Println("Error updating user room:", err)
			return false, err
		}
		_, err = tx.Exec("UPDATE rooms SET opened = 0 WHERE id =?", roomId)
		if err != nil {
			fmt.Println("Error updating room opened:", err)
			return false, err
		}
		return true, nil
	}
}

// 计分
func AddRecord(roomId int, fromUser string, toUser string, score int) error {
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
	// 查询 fromUser 和 toUser 的 score
	var fromScore, toScore model.Score
	err = tx.QueryRow("SELECT * FROM scores WHERE openid =? AND roomId =?", fromUser, roomId).Scan(&fromScore.Id, &fromScore.Openid, &fromScore.RoomId, &fromScore.Score, &fromScore.CreateData)
	if err != nil {
		fmt.Println("Error querying fromUser score:", err)
		return err
	}
	err = tx.QueryRow("SELECT * FROM scores WHERE openid =? AND roomId =?", toUser, roomId).Scan(&toScore.Id, &toScore.Openid, &toScore.RoomId, &toScore.Score, &toScore.CreateData)
	if err != nil {
		fmt.Println("Error querying toUser score:", err)
		return err
	}
	// 更新 fromUser 和 toUser 的 score
	_, err = tx.Exec("UPDATE scores SET score =? WHERE openid =? AND roomId =?", fromScore.Score-score, fromUser, roomId)
	if err != nil {
		fmt.Println("Error updating fromUser score:", err)
		return err
	}
	_, err = tx.Exec("UPDATE scores SET score =? WHERE openid =? AND roomId =?", toScore.Score+score, toUser, roomId)
	if err != nil {
		fmt.Println("Error updating toUser score:", err)
		return err
	}
	// 插入记录
	_, err = tx.Exec("INSERT INTO records (roomId, score, fromUser, toUser, createData) VALUES (?,?,?,?, NOW())", roomId, score, fromUser, toUser)
	if err != nil {
		fmt.Println("Error inserting record:", err)
		return err
	}
	return nil
}

// 修改昵称
func UpdateNickname(openid string, nickname string) error {
	_, err := db.Exec("UPDATE users SET nickname =? WHERE openid =?", nickname, openid)
	return err
}

// 获取房间内所有用户积分
func GetRoomScores(roomId int) ([]model.Score, error) {
	var scores []model.Score
	rows, err := db.Query("SELECT * FROM scores WHERE roomId =?", roomId)
	if err != nil {
		fmt.Println("Error querying room scores:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var score model.Score
		err = rows.Scan(&score.Id, &score.Openid, &score.RoomId, &score.Score, &score.CreateData)
		if err != nil {
			fmt.Println("Error scanning room scores:", err)
			return nil, err
		}
		scores = append(scores, score)
	}
	return scores, nil
}
