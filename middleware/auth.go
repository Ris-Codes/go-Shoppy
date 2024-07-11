package middleware

import (
	"fmt"

	"net/http"
	"os"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func UserAuth(c *gin.Context) {
	//Get the cookie off req
	tokenString, err := c.Cookie("UserAuthorization")

	if err != nil {
		c.Redirect(http.StatusFound, "/user/login")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Decode/validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		c.JSON(500, gin.H{
			"Status": "False",
			"Error":  "Error occured while token genaration",
		})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Set("userid", claims["sub"])

		c.Next()

	} else {
		c.AbortWithStatus(http.StatusUnauthorized)

	}
}

func AdminAuth(c *gin.Context) {
	//Get the cookie off req
	tokenString, err := c.Cookie("AdminAuthorization")

	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login")

		c.AbortWithStatus(http.StatusUnauthorized)
	}

	// Decode/validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		c.JSON(401, gin.H{
			"Message": "Admin logout",
		})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the exp

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.JSON(401, gin.H{
				"message": "Token expired",
			})
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Atach to req
		c.Set("adminid", claims["sub"])

		// continuew
		c.Next()

	} else {
		c.JSON(401, gin.H{
			"message": "Invalid token",
		})
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
