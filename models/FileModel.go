// models/file.go
package models

import "time"

type FileMetadata struct {
	UserID     string    `json:"user_id" bson:"user_id"`
	FileName   string    `json:"file_name" bson:"file_name"`
	UploadDate time.Time `json:"upload_date" bson:"upload_date"`
	FileSize   int64     `json:"file_size" bson:"file_size"`
	URL        string    `json:"url" bson:"url"`
}
