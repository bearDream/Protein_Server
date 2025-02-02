package controllers

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"Protein_Server/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var user models.User
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter error"})
		return
	}
	if err := database.Database.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-mail already exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Registered successfully"})
}

func LogIn(c *gin.Context) {
	var user models.User
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter error"})
		return
	}
	var count int64
	if err := database.Database.Model(&models.User{}).Where("email=? and password=?", user.Email, user.Password).Count(&count).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The account or password is incorrect"})
		return
	}
	// Return token
	var claims services.AccountClaims
	claims.User = user
	token := services.GenerateToken(&claims)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func ForgetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

func SendCode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

func GetUserInformation(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	var user models.User
	if err := database.Database.Where("email = ?", userByToken.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failure to obtain information"})
		return
	}
	// Processing the resulting data
	result := map[string]interface{}{
		"ID":        user.ID,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
		"Email":     user.Email,
		"NewCount":  user.NewCount,
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func GetOtherUser(c *gin.Context) {
	// Myself email
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	// Look for emails other than myself
	var users []models.User
	if err := database.Database.Where("email <> ?", userByToken.Email).Find(&users).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Processing the resulting data
	results := make([]map[string]interface{}, 0)
	for _, user := range users {
		results = append(results, map[string]interface{}{
			"ID":    user.ID,
			"Email": user.Email,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": results})
}
