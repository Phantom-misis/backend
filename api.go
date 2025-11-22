package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func login(c *gin.Context) {
	var form LoginForm
	if c.ShouldBind(&form) == nil {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Hello %s", form.User)})
	}
}
