package controller

import (
	"context"
	"golang-restaurant-backend-app/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := userCollection.Find(context.TODO(), bson.M{})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the users"})
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, allUsers)
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
