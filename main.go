package main

import (
	"scoringMP/config"
	"scoringMP/routers"
	"scoringMP/service/db"

	"github.com/gin-gonic/gin"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		return
	}
	err = db.InitDB()
	if err != nil {
		return
	}
	err = db.CreateTables()
	if err != nil {
		return
	}
	db.QueryUser("112323")
	r := gin.Default()
	routers.InitRouter(r)
	r.Run(config.Config.Port)
}
