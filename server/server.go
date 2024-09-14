package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"trademarkia/handlers"

	"trademarkia/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	handlers.ConnectToPostgres()
	handlers.ConnectToMongoDB()
	defer handlers.DisconnectFromMongoDB()
	defer handlers.DisconnectFromPostgres()

	PORT := config.PORT
	app := fiber.New()

	app.Use(logger.New())

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, world!",
		})
	})
	app.Post("/register", handlers.SignupHandler)
	app.Post("/login", handlers.LoginHandler)
	// app.Post("/upload", handlers.UploadFile)
	go func() {
		err := app.Listen(":" + PORT)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	handlers.DisconnectFromMongoDB()
	handlers.DisconnectFromPostgres()
}
