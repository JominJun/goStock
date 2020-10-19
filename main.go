package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
)

func main() {
	var r *gin.Engine
	r = gin.Default()

  r.LoadHTMLGlob("templates/*")
  
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
  })
  
	r.Run(":8080")
}