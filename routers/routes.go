package routers

import (
	"scoringMP/handles"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/login", handles.Login)
		api.GET("/userRoom", handles.GetUserRoom)
		api.GET("/history", handles.GetHistory)
		api.POST("/room", handles.CreateRoom)
		api.POST("/joinRoom", handles.JoinRoom)
		api.GET("/room", handles.GetRoomDetail)
		api.POST("/record", handles.AddRecord)
		api.PUT("/nickname", handles.UpdateNickname)
		api.DELETE("/room", handles.ExitRoom)
	}
}
