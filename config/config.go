package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	SECRET_KEY         = os.Getenv("SECRET_KEY")
	PORT        string = os.Getenv("PORT")
	MONGO_URL   string = os.Getenv("MONGO_URL")
	PG_HOST            = os.Getenv("PG_HOST")
	PG_PORT            = os.Getenv("PG_PORT")
	PG_USER            = os.Getenv("PG_USER")
	PG_PASSWORD        = os.Getenv("PG_PASSWORD")
	PG_DBNAME          = os.Getenv("PG_DBNAME")
)
