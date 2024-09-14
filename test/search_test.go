package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"trademarkia/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSearchFilesHandler(t *testing.T) {
	// Setup Fiber app
	app := fiber.New()

	// Register route with SearchFilesHandler
	app.Get("/search", func(c *fiber.Ctx) error {
		// Mock userID in the context
		c.Locals("userID", "valid User_ID") // Replace this with a valid test user ID
		return handlers.SearchFilesHandler(c)
	})

	// Create test request to search files
	req := httptest.NewRequest(http.MethodGet, "/search?name=check.jpg&limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt-token") // Mocking a valid JWT token for authenticated request
	// Create a test response recorder
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send test request: %v", err)
	}

	// Assert that the status code is 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Optionally: Add more assertions based on the response body or headers
}
