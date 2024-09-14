package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"trademarkia/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

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
	_, err := S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
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

func SearchFilesHandler(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	name := c.Query("name")
	date := c.Query("date")
	limit := c.QueryInt("limit", 10)  // Default limit
	offset := c.QueryInt("offset", 0) // Default offset

	// Format cache key
	cacheKey := fmt.Sprintf("files:%s:%s:%s:%d:%d", userID, name, date, limit, offset)
	fmt.Printf("Cache Key: %s\n", cacheKey)

	// Attempt to retrieve from cache
	cachedData, err := RedisClient.Get(context.Background(), cacheKey).Result()
	if err != redis.Nil {
		fmt.Println("Cache miss, querying database...")

		// Prepare SQL query
		var query string
		params := []interface{}{userID}
		paramIndex := 2

		query = `SELECT file_id, filename, upload_date, s3_url FROM files WHERE user_id = $1`

		if name != "" {
			query += fmt.Sprintf(` AND filename ILIKE $%d`, paramIndex)
			params = append(params, "%"+name+"%")
			paramIndex++
		}
		if date != "" {
			query += fmt.Sprintf(` AND upload_date::date = $%d`, paramIndex)
			params = append(params, date)
			paramIndex++
		}
		if limit > 0 {
			query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, paramIndex, paramIndex+1)
			params = append(params, limit, offset)
		}

		// Connect to database
		conn, err := pgx.Connect(context.Background(), getPostgresURL())
		if err != nil {
			log.Println("Database Connection Error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection error"})
		}
		defer conn.Close(context.Background())

		// Execute query
		rows, err := conn.Query(context.Background(), query, params...)
		if err != nil {
			log.Println("Database Query Error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query error"})
		}
		defer rows.Close()

		// Collect results
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

		// Cache result
		filesJSON, _ := json.Marshal(files)
		err = RedisClient.Set(context.Background(), cacheKey, filesJSON, 5*time.Minute).Err()
		if err != nil {
			log.Println("Error setting cache:", err)
		} else {
			fmt.Println("Data cached successfully.")
		}

		return c.Status(fiber.StatusOK).JSON(files)
	} else if err != nil {
		log.Println("Redis Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Redis error"})
	} else {
		fmt.Println("Cache hit, returning cached data...")
		var files []fiber.Map
		err = json.Unmarshal([]byte(cachedData), &files)
		if err != nil {
			log.Println("Error unmarshalling cache data:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cache data error"})
		}
		return c.Status(fiber.StatusOK).JSON(files)
	}
}

func UpdateFileMetadataHandler(c *fiber.Ctx) error {
	fileID := c.Params("file_id")
	newName := c.Query("name")

	// Update metadata in the database
	conn, err := pgx.Connect(context.Background(), getPostgresURL())
	if err != nil {
		log.Println("Database Connection Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection error"})
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `UPDATE files SET filename = $1 WHERE file_id = $2`, newName, fileID)
	if err != nil {
		log.Println("Database Update Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update file metadata"})
	}

	// Invalidate the cache
	cacheKey := fmt.Sprintf("files:*:*:*:*:*") // Pattern to invalidate all cache for this user
	keys, err := RedisClient.Keys(context.Background(), cacheKey).Result()
	if err == nil {
		for _, key := range keys {
			RedisClient.Del(context.Background(), key)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "File metadata updated successfully"})
}
