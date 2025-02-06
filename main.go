package main

import (
	"scoringMP/config"
	"scoringMP/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routers.InitRouter(r)
	r.Run(config.Port)
}
