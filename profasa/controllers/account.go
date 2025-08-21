package controllers

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"Protein_Server/services"
	"Protein_Server/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Register(c *gin.Context) {
	var user models.User
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&user); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}
	if err := database.Database.Create(&user).Error; err != nil {
		utils.Error(c, 400, "E-mail already exists")
		return
	}
	utils.Success(c, nil, "Registered successfully")
}

func LogIn(c *gin.Context) {
	var loginRequest models.User
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&loginRequest); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}
	
	// 查找匹配邮箱和密码的用户
	var users []models.User
	if err := database.Database.Where("email = ? AND password = ?", loginRequest.Email, loginRequest.Password).Find(&users).Error; err != nil {
		utils.Error(c, 500, "Network error.")
		return
	}
	
	// 用户不存在
	if len(users) == 0 {
		utils.Error(c, 400, "Invalid email and/or password. Please try again.")
		return
	}
	
	// 用户存在，生成 token
	user := users[0]
	var claims services.AccountClaims
	claims.User = user
	token := services.GenerateToken(&claims)
	
	// 返回与 Node.js 版本一致的数据结构
	utils.Success(c, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"token": token,
	}, "login success")
}

func ForgetPassword(c *gin.Context) {
	utils.Success(c, nil, "Created successfully")
}

func GetUserInformation(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	var user models.User
	if err := database.Database.Where("email = ?", userByToken.Email).First(&user).Error; err != nil {
		utils.Error(c, 400, "Failure to obtain information")
		return
	}
	// Processing the resulting data
	result := map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"newCount": user.NewCount,
	}
	utils.Success(c, result, "ok")
}

func GetOtherUser(c *gin.Context) {
	// Myself email
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	// Look for emails other than myself
	var users []models.User
	if err := database.Database.Where("email <> ?", userByToken.Email).Find(&users).Error; err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	// Processing the resulting data
	results := make([]map[string]interface{}, 0)
	for _, user := range users {
		results = append(results, map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		})
	}
	utils.Success(c, results, "ok")
}

func GetAllUserNotMe(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	var users []models.User
	if err := database.Database.Where("id <> ?", userByToken.ID).Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Network error.")
		return
	}
	if len(users) == 0 {
		utils.Success(c, "error", "ok")
		return
	}
	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		})
	}
	utils.Success(c, result, "ok")
}
