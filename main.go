package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/static"
)

func main() {
	var r *gin.Engine
	r = gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./client/build", true)))
	r.Run(":8080")
}