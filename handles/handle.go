package handles

import (
	"database/sql"
	"fmt"
	"scoringMP/service/db"
	"scoringMP/service/mp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginModel struct {
	Code string `json:"code"`
}

// 登录
func Login(c *gin.Context) {
	var data LoginModel
	err := c.Bind(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": "body error"})
		return
	}
	code := data.Code
	if code == "" {
		c.JSON(400, gin.H{"error": "code is required"})
		return
	}
	openId, err := mp.Code2Session(code)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
	}
	// 查询是否注册
	_, err = db.QueryUser(openId)
	if err != nil {
		if err == sql.ErrNoRows {
			// 注册，昵称为用户openId前6位
			err = db.RegisterUser(openId, "用户"+openId[:6])
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"openId": openId, "roomId": nil})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 查询用户房间
	room, err := db.QueryUserRoom(openId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(200, gin.H{"openId": openId, "roomId": nil})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"openId": openId, "roomId": room.Id})
}

// 获取用户房间
func GetUserRoom(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	room, err := db.QueryUserRoom(openId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Done()
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
	}
	c.String(200, fmt.Sprint(room.Id))
}

// 获取用户历史战绩
func GetHistory(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	scores, err := db.QueryHistory(openId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, scores)
}

type JoinRoomModel struct {
	RoomId int `json:"roomId"`
}

// 加入房间
func JoinRoom(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	var data JoinRoomModel
	err := c.Bind(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": "body error"})
		return
	}
	err = db.JoinRoom(openId, data.RoomId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(400, gin.H{"error": "room is not exist"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
}

// 创建/回到房间
func CreateRoom(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	roomId, err := db.CreateRoom(openId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.String(200, fmt.Sprint(roomId))
}

// 获取用户列表和积分列表
func GetRoomDetail(c *gin.Context) {
	roomId, err := strconv.Atoi(c.Query("roomId"))
	if err != nil {
		c.JSON(400, gin.H{"error": "roomId is required"})
		return
	}
	opened, err := db.CheckRoom(roomId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	users, err := db.GetRoomUsers(roomId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	records, err := db.GetRoomRecords(roomId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"users": users, "records": records, "isOpen": opened})
}

type AddRecordModel struct {
	RoomId   int    `json:"roomId"`
	Score    int    `json:"score"`
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
}

// 计分
func AddRecord(c *gin.Context) {
	var data AddRecordModel
	err := c.Bind(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 判断房间是否关闭
	opened, err := db.CheckRoom(data.RoomId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !opened {
		c.JSON(400, gin.H{"error": "room is not opened"})
		return
	}
	// 判断 fromUser 和 toUser 是否在房间中
	fromUser, err := db.QueryUser(data.FromUser)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !fromUser.RoomId.Valid || fromUser.RoomId.Int64 != int64(data.RoomId) {
		c.JSON(400, gin.H{"error": "fromUser is not in room"})
		return
	}
	toUser, err := db.QueryUser(data.ToUser)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !toUser.RoomId.Valid || toUser.RoomId.Int64 != int64(data.RoomId) {
		c.JSON(400, gin.H{"error": "toUser is not in room"})
		return
	}
	// 插入记录
	err = db.AddRecord(data.RoomId, data.FromUser, data.ToUser, data.Score)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.String(200, "ok")
}

type ModifyNicknameModel struct {
	Nickname string `json:"nickname"`
}

// 修改昵称
func UpdateNickname(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	var data ModifyNicknameModel
	err := c.Bind(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": "body error"})
		return
	}
	err = db.UpdateNickname(openId, data.Nickname)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.String(200, "ok")
}

type ExitRoomModel struct {
	RoomId int `json:"roomId"`
}

// 退出房间
func ExitRoom(c *gin.Context) {
	openId := c.Request.Header.Get("openId")
	var data ExitRoomModel
	err := c.Bind(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": "body error"})
		return
	}
	_, err = db.QuitRoom(openId, data.RoomId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
}
