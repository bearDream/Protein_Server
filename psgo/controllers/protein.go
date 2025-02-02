package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetInformation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

func MaxSize(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}
