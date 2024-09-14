package middlewares

import (
	"fmt"
	"strings"
	"trademarkia/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware is a middleware that verifies JWT tokens
func AuthMiddleware(c *fiber.Ctx) error {
	// Extract token from the Authorization header
	tokenHeader := c.Get("Authorization")
	if tokenHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	// Split the token header to get the token part
	tokenParts := strings.Split(tokenHeader, "Bearer ")
	if len(tokenParts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token format",
		})
	}
	tokenString := tokenParts[1]

	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the token signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(config.SECRET_KEY), nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	// Extract user ID from claims
	userID, ok := claims["_id"].(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in token",
		})
	}

	// Store user ID in context
	c.Locals("userID", userID)

	return c.Next()
}

// ExtractUserID extracts the user ID from the context
func ExtractUserID(c *fiber.Ctx) (string, error) {
	// Retrieve user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}

	return userID, nil
}
