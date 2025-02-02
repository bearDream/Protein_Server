package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RenderPDB2OBJ(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}
