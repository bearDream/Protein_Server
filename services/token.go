package services

import (
	"Protein_Server/models"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type AccountClaims struct {
	models.User
	jwt.StandardClaims
}

var (
	secret     = []byte("19960609")
	effectTime = 12 * time.Hour
)

func GenerateToken(account *AccountClaims) string {
	account.ExpiresAt = time.Now().Add(effectTime).Unix()
	sign, err := jwt.NewWithClaims(jwt.SigningMethodHS256, account).SignedString(secret)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	// sign is a token string
	return sign
}

func JwtVerify(c *gin.Context) {
	// get token from header
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token not exist!"})
		c.Abort()
		return
	}
	c.Set("account", parseToken(token, c))
}

func parseToken(tokenString string, c *gin.Context) *AccountClaims {
	token, err := jwt.ParseWithClaims(tokenString, &AccountClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token error!"})
		c.Abort()
	}
	claims := token.Claims.(*AccountClaims)
	return claims
}

func Refresh(tokenString string) string {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &AccountClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	claims := token.Claims.(*AccountClaims)
	jwt.TimeFunc = time.Now
	claims.StandardClaims.ExpiresAt = time.Now().Add(2 * time.Hour).Unix()
	return GenerateToken(claims)
}
