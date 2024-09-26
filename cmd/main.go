package main

import (
	"log"

	api "server/internal/api"
	"server/internal/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// echo middleware
	e.Use(middleware.RateLimiter)

	// register API routes and middleware
	api.RegisterRoutes(e)

	// start the echo server
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
