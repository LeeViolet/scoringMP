package handles

import (
	mp "scoringMP/service"

	"github.com/gin-gonic/gin"
)

type SUser struct {
	Code     string `json:"code"`
	Nickname string `json:"nickname"`
}

func Login(c *gin.Context) {
	var user SUser
	err := c.Bind(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "body error"})
		return
	}
	code := user.Code
	if code == "" {
		c.JSON(400, gin.H{"error": "code is required"})
		return
	}
	openId, err := mp.Code2Session(code)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
	}

	c.String(200, openId)
}
