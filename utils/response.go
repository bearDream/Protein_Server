package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Success(c *gin.Context, data interface{}, msg string) {
	c.JSON(200, Response{Code: 200, Data: data, Message: msg})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(200, Response{Code: code, Data: nil, Message: msg})
} 