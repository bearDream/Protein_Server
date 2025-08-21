package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Success(c *gin.Context, data interface{}, msg string) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    data,
		"message": msg,
	})
}

func Error(c *gin.Context, code int, msg string) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})
}
