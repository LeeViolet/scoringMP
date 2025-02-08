package handles

import (
	"database/sql"
	"scoringMP/service/db"
	"scoringMP/service/mp"

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
