package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/contrib/static"
)

func main() {
	var r *gin.Engine
	r = gin.Default()

  r.Use(static.Serve("/", static.LocalFile("./client/build", true)))
  
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
  })
  
	r.Run(":8080")
}