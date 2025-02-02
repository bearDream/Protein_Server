package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func FoldDurationList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

func ParameterList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}
func ParameterExport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}
