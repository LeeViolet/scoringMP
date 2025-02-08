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
			// 注册，昵称为随机数字
			err = db.RegisterUser(openId, "user"+openId)
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.String(200, openId)
}

// 获取用户房间 id
func GetUserRoom(c *gin.Context) {
	openId := c.Query("openId")
	roomId, err := db.QueryUserRoom(openId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if roomId.Valid {
		c.String(200, fmt.Sprint(roomId.Int64))
	} else {
		c.String(204, "")
	}
}

// 获取用户历史战绩
func GetHistory(c *gin.Context) {
	openId := c.Query("openId")
	scores, err := db.QueryHistory(openId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, scores)
}

// 创建房间
func CreateRoom(c *gin.Context) {
	openId := c.Query("openId")
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
	if !opened {
		c.JSON(400, gin.H{"error": "room is not opened"})
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
	c.JSON(200, gin.H{"users": users, "records": records})
}
