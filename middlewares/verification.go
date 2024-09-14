// middlewares/middlewares.go
package middlewares

import (
	"fmt"
	"trademarkia/config"

	"github.com/gofiber/fiber/v2"
	jwt "github.com/gofiber/jwt/v3"
)

// AuthMiddleware returns a JWT middleware handler
func AuthMiddleware() fiber.Handler {
	return jwt.New(jwt.Config{
		SigningKey: []byte(config.SECRET_KEY),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		},
	})
}

// ExtractUserID extracts the user ID from the JWT token in the request context
func ExtractUserID(c *fiber.Ctx) (string, error) {
	// Extract the JWT claims from the context
	claims, ok := c.Locals("user").(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Retrieve user ID from claims
	userID, ok := claims["_id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in token")
	}

	return userID, nil
}
