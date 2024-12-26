package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"server/config"
	"server/internal/db"
	"server/internal/middleware"
	"server/internal/server"
	"sync"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Set Gin to release mode for production
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	configValue := config.LoadConfig()
	diceDBAdminClient, err := db.InitDiceClient(configValue, true)
	if err != nil {
		slog.Error("Failed to initialize DiceDB Admin client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	diceDBClient, err := db.InitDiceClient(configValue, false)
	if err != nil {
		slog.Error("Failed to initialize DiceDB client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	// Register a cleanup manager, this runs user DiceDB instance cleanup job at configured frequency
	cleanupManager := server.NewCleanupManager(diceDBAdminClient, diceDBClient, configValue.Server.CronCleanupFrequency)
	wg.Add(1)
	go cleanupManager.Run(ctx, &wg)

	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := configValue.Server.AllowedOrigins
		isAllowed := false

		if len(allowedOrigins) == 0 {
			// Allow all origins, if none are specified in config
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				http.Error(c.Writer, "403 - Forbidden", http.StatusForbidden)
				c.Abort()
				return
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			if len(allowedOrigins) == 0 || isAllowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})
	router.Use(middleware.TrailingSlashMiddleware)
	router.Use((middleware.NewRateLimiterMiddleware(diceDBAdminClient,
		configValue.Server.RequestLimitPerMin,
		configValue.Server.RequestWindowSec,
	).Exec))

	httpServer := server.NewHTTPServer(
		router,
		diceDBAdminClient,
		diceDBClient,
		configValue.Server.RequestLimitPerMin,
		configValue.Server.RequestWindowSec,
	)

	// Register routes
	router.GET("/health", gin.WrapF(httpServer.HealthCheck))
	router.POST("/shell/exec/:cmd", gin.WrapF(httpServer.CliHandler))
	router.GET("/search", gin.WrapF(httpServer.SearchHandler))

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run the HTTP Server
		if err := httpServer.Run(context.Background()); err != nil {
			slog.Error("server failed: %v\n", slog.Any("err", err))
			diceDBAdminClient.CloseDiceDB()
			cancel()
		}
	}()

	wg.Wait()
	slog.Info("Server has shut down gracefully")
}
