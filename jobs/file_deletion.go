package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"trademarkia/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	deleteInterval = 24 * time.Hour
)

func StartFileDeletionJob(db *sql.DB, s3Client *s3.Client) {
	go func() {
		for {
			deleteExpiredFiles(db, s3Client)
			time.Sleep(deleteInterval)
		}
	}()
}

func deleteExpiredFiles(db *sql.DB, s3Client *s3.Client) {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT file_id, s3_url FROM files WHERE upload_date < NOW() - INTERVAL '3 days'`)
	if err != nil {
		log.Println("Database Query Error:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var fileID, s3URL string
		if err := rows.Scan(&fileID, &s3URL); err != nil {
			log.Println("Database Scan Error:", err)
			return
		}
		_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(config.S3_BUCKET),
			Key:    aws.String(fileID),
		})
		if err != nil {
			log.Println("S3 Delete Error:", err)
			continue
		}
		_, err = db.ExecContext(ctx, `DELETE FROM files WHERE file_id = $1`, fileID)
		if err != nil {
			log.Println("Database Deletion Error:", err)
		}
	}
}

func deleteFileFromS3(s3Client *s3.Client, s3URL string) error {
	// Extract the bucket and key from the s3URL
	bucket, key := extractBucketAndKey(s3URL)

	// Create the delete object input
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Call S3 to delete the object
	_, err := s3Client.DeleteObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

func extractBucketAndKey(s3URL string) (string, string) {
	parts := strings.Split(s3URL, "/")
	bucket := parts[2]
	key := parts[3]

	return bucket, key
}
