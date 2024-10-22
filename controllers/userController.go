package controller

import (
	"context"
	"golang-restaurant-backend-app/database"
	helper "golang-restaurant-backend-app/helper"
	"golang-restaurant-backend-app/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Default pagination values
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		// Calculate start index for pagination
		startIndex := (page - 1) * recordPerPage

		// MongoDB aggregation pipeline stages
		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
		skipStage := bson.D{{Key: "$skip", Value: startIndex}}
		limitStage := bson.D{{Key: "$limit", Value: recordPerPage}}
		projectStage := bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0}, // Exclude MongoDB's default "_id" field
			{Key: "email", Value: 1},
			{Key: "first_name", Value: 1},
			{Key: "last_name", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}}}

		// Perform aggregation
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, skipStage, limitStage, projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
			return // Exit early if aggregation fails
		}

		// Decode aggregation results into a slice of maps
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return response
		if len(allUsers) > 0 {
			c.JSON(http.StatusOK, allUsers)
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "No users found"})
		}
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		userId := c.Param("user_id")

		var user models.Table

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the user"})
		}
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		// Bind and validate user input
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validatorErr := validate.Struct(user)
		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}

		// Check if email or phone number already exists
		emailOrPhoneExists, err := userCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"phone": user.Phone},
			},
		})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking for existing user"})
			return
		}
		if emailOrPhoneExists > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or phone number already exists"})
			return
		}

		// Hash the password
		password := HashPassword(*user.Password)
		user.Password = &password

		// Set user metadata
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		// Generate JWT tokens
		token, refreshToken, tokenErr := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		if tokenErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		user.Token = &token
		user.Refresh_Token = &refreshToken

		// Insert user into database
		result, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User was not created"})
			return
		}

		// Return successful response
		c.JSON(http.StatusOK, result)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure the context is cancelled after the function executes

		var user models.User
		var foundUser models.User

		// Bind and validate user input
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find user by email
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Verify password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		// Generate tokens
		token, refreshToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		// Update tokens in the database
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// Exclude sensitive information before returning the user data
		foundUser.Password = nil

		// Return the user and tokens
		c.JSON(http.StatusOK, gin.H{
			"user":         foundUser,
			"token":        token,
			"refreshToken": refreshToken,
		})
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Println(err)
	}
	return string(bytes)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		check = false
	}
	return check, msg
}
