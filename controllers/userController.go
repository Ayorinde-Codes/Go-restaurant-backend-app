package controller

import (
	"github.com/gin-gonic/gin"
)

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// id := c.Param("id")
		// todo: get user by id
	}
}
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo: create new user
	}
}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo: login user
	}
}

func HashPassword(password string) string {
	// todo: hash password
}

func VerifyPassword(usePassword string, providePassword string) (bool, string) {

}
