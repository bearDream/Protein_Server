package database

import (
	"Protein_Server/logger"
	"Protein_Server/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB

// Connect to database
func init() {
	dsn := "root:TxMysql$100*!@tcp(101.35.87.147:3306)/protein_new?charset=utf8&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		logger.Error("数据库连接失败: %v", err)
		return
	}
	// Automatically build table
	database.AutoMigrate(&models.AlphaFoldQueue{}, &models.ESMQueue{}, &models.ITasserQueue{}, &models.Note{}, &models.ProteinInformation{}, &models.Share{}, &models.Task{}, &models.User{})
	Database = database
}
