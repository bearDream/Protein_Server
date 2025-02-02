package database

import (
	"Protein_Server/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB

// Connect to database
func init() {
	dsn := "root:miyu960609@/protein?charset=utf8&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Automatically build table
	database.AutoMigrate(&models.AlphaFoldQueue{}, &models.ITasserQueue{}, &models.Note{}, &models.ProteinInformation{}, &models.Share{}, &models.Task{}, &models.User{})
	Database = database
}
