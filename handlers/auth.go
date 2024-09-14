// handlers/auth.go
package handlers

import (
	"context"
	"log"
	"time"
	"trademarkia/config"
	"trademarkia/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func SignupHandler(c *fiber.Ctx) error {
	var user models.SignupUser
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	existingUser := bson.M{"email": user.Email}
	count, err := collection.CountDocuments(context.Background(), existingUser)
	if err != nil {
		log.Fatal(err)
	}

	existingUsername := bson.M{"username": user.Username}
	countuser, err := collection.CountDocuments(context.Background(), existingUsername)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User with this email already exists",
		})
	}

	if countuser > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User with this username already exists",
		})
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt password",
		})
	}
	newUser := bson.M{
		"email":    user.Email,
		"username": user.Username,
		"password": string(hashedPassword),
	}

	_, err = collection.InsertOne(context.Background(), newUser)
	if err != nil {
		log.Fatal(err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Signed up successfully!",
	})
}

func LoginHandler(c *fiber.Ctx) error {
	var loginCredentials models.LoginUser
	if err := c.BodyParser(&loginCredentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	storedPassword, userID := getPasswordAndIDFromDatabase(loginCredentials.Email)

	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginCredentials.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials",
		})
	}

	token := generateToken(userID.Hex())

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful!",
		"token":   token,
	})
}

func getPasswordAndIDFromDatabase(email string) (string, primitive.ObjectID) {
	var result bson.M
	filter := bson.M{"email": email}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	hashedPassword := result["password"].(string)
	userID := result["_id"].(primitive.ObjectID)
	return hashedPassword, userID
}

func generateToken(userID string) string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Add expiration time

	tokenString, err := token.SignedString([]byte(config.SECRET_KEY))
	if err != nil {
		log.Fatal(err)
	}

	return tokenString
}
