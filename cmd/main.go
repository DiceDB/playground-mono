package main

import (
	"log"

	"server/config"
	api "server/internal/api"
	"server/internal/db"
	"server/internal/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	db.InitializeDiceDB()
	config.LoadConfig()
	defer db.CloseDiceDB()

	// echo middleware
	e.Use(middleware.RateLimiter)

	// Register API routes and middleware
	api.RegisterRoutes(e)

	// Start the server
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
