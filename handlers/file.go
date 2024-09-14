package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"trademarkia/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

var s3Client *s3.Client

func init() {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(config.AWS_REGION))
	if err != nil {
		log.Fatal(err)
	}

	s3Client = s3.NewFromConfig(cfg)
}

func UploadHandler(c *fiber.Ctx) error {
	log.Println("UploadHandler called")

	// Extract userID from JWT token claim
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Println("Failed to get file:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to get file"})
	}

	fileContent, err := file.Open()
	if err != nil {
		log.Println("Failed to open file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer fileContent.Close()

	fileID := uuid.New().String()
	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", config.S3_BUCKET, config.AWS_REGION, fileID)

	chunkSize := 10 * 1024 * 1024 // 10 MB
	var wg sync.WaitGroup
	chunkChannel := make(chan []byte, 10)
	errorChannel := make(chan error, 1)

	go func() {
		defer close(chunkChannel)
		buffer := make([]byte, chunkSize)
		for {
			n, err := fileContent.Read(buffer)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buffer[:n])
				chunkChannel <- chunk
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				errorChannel <- err
				return
			}
		}
	}()

	for chunk := range chunkChannel {
		wg.Add(1)
		go func(data []byte) {
			defer wg.Done()
			if err := uploadChunkToS3(data, fileID); err != nil {
				errorChannel <- err
			}
		}(chunk)
	}

	wg.Wait()
	close(errorChannel)

	if err, ok := <-errorChannel; ok {
		log.Println("Failed to upload to S3:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload to S3"})
	}

	if err := saveFileMetadata(file.Filename, fileID, s3URL, userID); err != nil {
		log.Println("Failed to save metadata:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save metadata"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"url": s3URL})
}

func uploadChunkToS3(chunk []byte, fileID string) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(config.S3_BUCKET),
		Key:    aws.String(fileID),
		Body:   bytes.NewReader(chunk),
	})
	return err
}

func saveFileMetadata(filename, fileID, s3URL, userID string) error {
	conn, err := pgx.Connect(context.Background(), getPostgresURL())
	if err != nil {
		log.Println("Database Connection Error:", err)
		return err
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `INSERT INTO files (file_id, filename, upload_date, s3_url, user_id) VALUES ($1, $2, $3, $4, $5)`,
		fileID, filename, time.Now(), s3URL, userID)
	return err
}

func getPostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.PG_USER,
		config.PG_PASSWORD,
		config.PG_HOST,
		config.PG_PORT,
		config.PG_DBNAME,
	)
}

func GetFilesHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	conn, err := pgx.Connect(context.Background(), getPostgresURL())
	if err != nil {
		log.Println("Database Connection Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection error"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `SELECT file_id, filename, upload_date, s3_url FROM files WHERE user_id = $1`, userID)
	if err != nil {
		log.Println("Database Query Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query error"})
	}
	defer rows.Close()

	files := []fiber.Map{}
	for rows.Next() {
		var fileID, filename, s3URL string
		var uploadDate time.Time
		if err := rows.Scan(&fileID, &filename, &uploadDate, &s3URL); err != nil {
			log.Println("Database Scan Error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database scan error"})
		}
		files = append(files, fiber.Map{
			"file_id":     fileID,
			"filename":    filename,
			"upload_date": uploadDate,
			"s3_url":      s3URL,
		})
	}

	return c.Status(fiber.StatusOK).JSON(files)
}

func ShareFileHandler(c *fiber.Ctx) error {
	fileID := c.Params("file_id")

	// Retrieve metadata
	conn, err := pgx.Connect(context.Background(), getPostgresURL())
	if err != nil {
		log.Println("Database Connection Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect to database"})
	}
	defer conn.Close(context.Background())

	var objectKey string
	err = conn.QueryRow(context.Background(), `SELECT s3_url FROM files WHERE file_id = $1`, fileID).Scan(&objectKey)
	if err != nil {
		log.Println("Database Query Error:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"share_url": objectKey})
}
