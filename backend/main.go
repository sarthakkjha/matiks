// Matiks Leaderboard Backend
// A high-performance, scalable leaderboard API built with Go and Gin.
//
// ARCHITECTURE FEATURES:
// 1. In-Memory Caching (cache/): O(1) reads, thread-safe, ready for Redis swap.
// 2. Ranking Engine (engine/): Snapshot-based O(1) rank lookups with tied-rank support.
// 3. Debounced Updates (services/): Batches hundreds of updates into single rebuilds.
// 4. Clean Architecture: Separated concerns (handlers -> services -> engine/cache -> db).
//
// Run with: go run main.go
// Environment: MONGODB_URI, PORT
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"matiks-leaderboard/database"
	"matiks-leaderboard/handlers"
	"matiks-leaderboard/services"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/matiks-leaderboard"
	}

	if err := database.Connect(ctx, mongoURI); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer database.Disconnect(context.Background())

	// Initialize leaderboard service (load cache, build snapshot)
	log.Println("📊 Initializing Leaderboard Service...")
	if err := services.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize service:", err)
	}

	// Seed database if empty
	count, err := services.SeedDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to seed database:", err)
	}
	if count > 0 {
		log.Printf("🌱 Seeded %d users\n", count)
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check (required for Render)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Matiks Leaderboard API",
			"version": "1.0.0",
			"docs":    "/api/stats",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Leaderboard
		api.GET("/leaderboard", handlers.GetLeaderboard)
		api.GET("/leaderboard/top/:n", handlers.GetTopN)

		// Users
		api.GET("/users/search", handlers.SearchUsers)
		api.GET("/users/:id", handlers.GetUserByID)
		api.POST("/users", handlers.CreateUser)
		api.PUT("/users/:id/score", handlers.UpdateScore)

		// Bulk updates (for demo)
		api.POST("/bulk-update/random", handlers.BulkUpdateRandom)
		api.POST("/bulk-update/value", handlers.BulkUpdateToValue)

		// Stats
		api.GET("/stats", handlers.GetStats)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Matiks Leaderboard API (Go)")
	log.Printf("📡 http://localhost:%s\n", port)
	log.Println("✅ Server ready!")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
