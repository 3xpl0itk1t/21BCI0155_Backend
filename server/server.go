package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"trademarkia/handlers"
	"trademarkia/jobs"
	"trademarkia/middlewares"

	"trademarkia/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// Connections
	handlers.ConnectToPostgres()
	handlers.ConnectToMongoDB()
	handlers.ConnectToS3()

	// Defer disconnecting from the connections when shutdown
	defer handlers.DisconnectFromMongoDB()
	defer handlers.DisconnectFromPostgres()
	defer handlers.DisconnectFromS3()

	go jobs.StartFileDeletionJob(handlers.PostgresDB, handlers.S3Client)

	PORT := config.PORT
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	// Public Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, world!",
		})
	})
	app.Post("/register", handlers.SignupHandler)
	app.Post("/login", handlers.LoginHandler)

	// Protected Routes
	protected := app.Group("/", middlewares.AuthMiddleware)

	protected.Post("/upload", handlers.UploadHandler)
	protected.Get("/files", handlers.GetFilesHandler)
	protected.Get("/share/:file_id", handlers.ShareFileHandler)
	protected.Get("/search", handlers.SearchFilesHandler)

	go func() {
		err := app.Listen(":" + PORT)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Disconnect from the connections
	handlers.DisconnectFromMongoDB()
	handlers.DisconnectFromPostgres()
	handlers.DisconnectFromS3()
}
