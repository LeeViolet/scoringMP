package model

import "database/sql"

type User struct {
	Openid     string        `json:"openid"`
	Nickname   string        `json:"nickname"`
	RoomId     sql.NullInt64 `json:"roomId"`
	CreateData string        `json:"createData"`
}

type Room struct {
	Id         int    `json:"id"`
	Owner      string `json:"owner"`
	CreateData string `json:"createData"`
	Opened     bool   `json:"opened"`
}

type Score struct {
	Id         int    `json:"id"`
	Openid     string `json:"openid"`
	RoomId     int    `json:"roomId"`
	Score      int    `json:"score"`
	CreateData string `json:"createData"`
}

type Record struct {
	Id         int    `json:"id"`
	RoomId     int    `json:"roomId"`
	Score      int    `json:"score"`
	FromUser   string `json:"fromUser"`
	ToUser     string `json:"toUser"`
	CreateData string `json:"createData"`
}
