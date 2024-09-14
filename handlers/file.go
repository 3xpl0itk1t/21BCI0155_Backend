// handlers/file.go
package handlers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"sync"
	"time"
	"trademarkia/middlewares"
	"trademarkia/models"

	"os"

	"github.com/gofiber/fiber/v2"
)

// UploadHandler handles file uploads
func UploadHandler(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse form",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No files uploaded",
		})
	}

	userID, err := middlewares.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	dir := "./uploads"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go processFile(file, userID, dir, &wg)
	}

	wg.Wait()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Files uploaded successfully!",
	})
}
func processFile(file *multipart.FileHeader, userID string, dir string, wg *sync.WaitGroup) {
	defer wg.Done()

	filePath := filepath.Join(dir, file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		log.Printf("Failed to save file: %s", err)
		return
	}

	fileMeta := models.FileMetadata{
		UserID:     userID,
		FileName:   file.Filename,
		UploadDate: time.Now(),
		FileSize:   file.Size,
		URL:        fmt.Sprintf("/uploads/%s", file.Filename),
	}

	_, err := collection.InsertOne(context.Background(), fileMeta)
	if err != nil {
		log.Printf("Failed to save file metadata: %s", err)
	}
}
