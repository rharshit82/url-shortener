package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rharshit82/url-shortner/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveUrl)
	app.Post("/api/v1", routes.ShortenUrl)
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Could not load environment variables")
	}
	var app *fiber.App = fiber.New()
	app.Use(logger.New())
	setupRoutes(app)
	app.Listen(os.Getenv("APP_PORT"))
}
