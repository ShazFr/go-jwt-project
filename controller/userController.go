package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ShazFR/go-jwt-project/database"
	"github.com/ShazFR/go-jwt-project/helper"
	"github.com/ShazFR/go-jwt-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUserInDB models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUserInDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
		}

		fmt.Println("Found user password (from DB):", *foundUserInDB.Password)
		fmt.Println("Password provided by user:", *user.Password)

		passwordIsValid, msg := VerifyPassword(*foundUserInDB.Password, *user.Password)

		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUserInDB.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUserInDB.Email, *foundUserInDB.FirstName, *foundUserInDB.LastName, *foundUserInDB.UserType, foundUserInDB.UserID)

		helper.UpdateAllTokens(token, refreshToken, foundUserInDB.UserID)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUserInDB.UserID}).Decode(&foundUserInDB)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, foundUserInDB)
	}
}

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)

}

func VerifyPassword(passwordFromDatabase string, providedPassword string) (bool, string) {

	fmt.Println("Stored password (hashed):", passwordFromDatabase)
	fmt.Println("Provided password:", providedPassword)

	err := bcrypt.CompareHashAndPassword([]byte(passwordFromDatabase), []byte(providedPassword))
	isPasswordCorrect := true
	msg := ""

	if err != nil {
		msg = "Password is Incorrect"
		isPasswordCorrect = false
	}
	return isPasswordCorrect, msg

}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ValidateUser(ctx, user, c)

		hashedPassword := HashPassword(*user.Password)
		user.Password = &hashedPassword

		now := time.Now().UTC()

		user.CreatedAt = now
		user.UpdatedAt = now
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		fmt.Print(user)

		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error when generating tokens"})
		}
		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "UserItem was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func ValidateUser(ctx context.Context, user models.User, c *gin.Context) {
	validationError := validate.Struct(user)
	if validationError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
	}

	emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Errr occured while checking signup validity email"})
	}

	phnoCount, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking signup validity phno"})
	}

	if emailCount > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Another user exists with this email"})
	}
	if phnoCount > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Another user exists with this phone number"})
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"userid": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordsPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordsPerPage < 1 {
			recordsPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordsPerPage
		// startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "null"},
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordsPerPage}},
				}},
			}},
		}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage, // redline here
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error Occured while aggregating the DB"})
			return
		}
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding the user list"})
			return
		}
		c.JSON(http.StatusOK, allUsers[0])

	}
}
