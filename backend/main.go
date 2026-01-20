// Matiks Leaderboard Backend
// A high-performance, scalable leaderboard API built with Go and Gin.
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
	godotenv.Load()

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

	log.Println("📊 Initializing Leaderboard Service...")
	if err := services.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize service:", err)
	}

	/*
		count, err := services.SeedDatabase(ctx)
		if err != nil {
			log.Fatal("Failed to seed database:", err)
		}
		if count > 0 {
			log.Printf("🌱 Seeded %d users\n", count)
		}
	*/

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Matiks Leaderboard API",
			"version": "1.0.0",
			"docs":    "/api/stats",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/leaderboard", handlers.GetLeaderboard)
		api.GET("/leaderboard/top/:n", handlers.GetTopN)

		api.GET("/users/search", handlers.SearchUsers)
		api.GET("/users/:id", handlers.GetUserByID)
		api.POST("/users", handlers.CreateUser)
		api.PUT("/users/:id/score", handlers.UpdateScore)

		api.POST("/bulk-update/random", handlers.BulkUpdateRandom)
		api.POST("/bulk-update/value", handlers.BulkUpdateToValue)

		api.GET("/stats", handlers.GetStats)
	}

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
