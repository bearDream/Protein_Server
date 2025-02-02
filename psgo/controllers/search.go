package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchByParameters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

func SearchByNatureLanguage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}
