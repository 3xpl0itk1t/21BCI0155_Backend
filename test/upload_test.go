package test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"trademarkia/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestUploadHandler(t *testing.T) {
	// Setup Fiber app
	app := fiber.New()

	// Register the route with UploadHandler
	app.Post("/upload", handlers.UploadHandler)

	// Create a test form file to simulate an upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "testfile.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("This is a test file content"))
	writer.Close()

	// Create test request
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer valid-jwt-token") // Mocking a valid JWT token for authenticated request

	// Create test response recorder
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send test request: %v", err)
	}

	// Assert that the status code is 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Optionally: Add more assertions based on the response body or headers
}
