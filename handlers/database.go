// handlers/database.go
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	"trademarkia/config"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var mongoClient *mongo.Client
var postgresDB *sql.DB

func ConnectToMongoDB() {
	connectionURI := config.MONGO_URL
	clientOptions := options.Client().ApplyURI(connectionURI)

	var err error
	mongoClient, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB Atlas!")

	collection = mongoClient.Database("Trademarkia").Collection("users")
}

func ConnectToPostgres() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PG_HOST, config.PG_PORT, config.PG_USER, config.PG_PASSWORD, config.PG_DBNAME)

	var err error
	postgresDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	err = postgresDB.Ping()
	if err != nil {
		log.Fatal("Failed to ping PostgreSQL:", err)
	}

	fmt.Println("Connected to PostgreSQL!")
}

func DisconnectFromMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := mongoClient.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Disconnected from MongoDB Atlas!")
}

func DisconnectFromPostgres() {
	err := postgresDB.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Disconnected from PostgreSQL!")
}
