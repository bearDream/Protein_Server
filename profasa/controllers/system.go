package controllers

import (
	"Protein_Server/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UploadFile(c *gin.Context) {

}

// GetQueueStatus 获取队列状态
func GetQueueStatus(c *gin.Context) {
	queueScheduler := services.GetGlobalQueueScheduler()
	status := queueScheduler.GetQueueStatus()
	
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "获取队列状态成功",
		"data": status,
	})
}
