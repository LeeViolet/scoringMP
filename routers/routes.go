package routers

import (
	"scoringMP/handles"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/login", handles.Login)
	}
}
