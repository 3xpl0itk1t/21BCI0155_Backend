package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	SECRET_KEY                   = os.Getenv("SECRET_KEY")
	PORT                  string = os.Getenv("PORT")
	MONGO_URL             string = os.Getenv("MONGO_URL")
	AWS_ACCESS_KEY_ID            = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY        = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS_REGION                   = os.Getenv("AWS_REGION")
	S3_BUCKET                    = os.Getenv("S3_BUCKET")
	PG_HOST                      = os.Getenv("PG_HOST")
	PG_PORT                      = os.Getenv("PG_PORT")
	PG_USER                      = os.Getenv("PG_USER")
	PG_PASSWORD                  = os.Getenv("PG_PASSWORD")
	PG_DBNAME                    = os.Getenv("PG_DBNAME")
)
