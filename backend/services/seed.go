// Package services contains the seeding logic for initial data.
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"matiks-leaderboard/database"
	"matiks-leaderboard/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedDatabase creates 11,000 users with proper rating distribution.
// - Ratings from 100 to 5000 (4,901 distinct values)
// - At least 2 users per rating
// - 3 users for lower ratings (100-1297) to reach 11,000 total
func SeedDatabase(ctx context.Context) (int, error) {
	collection := database.Collection("users")

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	if count >= 11000 {
		log.Printf("📊 Database already has %d users, skipping seed", count)
		return 0, nil
	}

	// Drop existing data to ensure clean seeding
	if count > 0 {
		log.Printf("🗑️ Dropping existing %d users for clean reseed...", count)
		if err := collection.Drop(ctx); err != nil {
			return 0, fmt.Errorf("failed to drop collection: %w", err)
		}
	}

	log.Println("🌱 Seeding 11,000 users...")

	var users []interface{}

	// Calculate: 4,901 ratings (100-5000) × 2 = 9,802 users
	// Remaining: 11,000 - 9,802 = 1,198 users
	// Add 3rd user for ratings 100-1297 (1,198 ratings)

	// First pass: 2 users per rating (100-5000)
	for rating := 5000; rating >= 100; rating-- {
		for i := 1; i <= 2; i++ {
			username := fmt.Sprintf("Player_%d_%d", rating, i)
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Score:    rating,
			})
		}
	}

	log.Printf("   Generated %d users (2 per rating)", len(users))

	// Second pass: Add 3rd user for lower ratings to reach 11,000
	// Need 1,198 more users (ratings 100-1297)
	extraNeeded := 11000 - len(users)
	extraAdded := 0

	for rating := 100; rating <= 1297 && extraAdded < extraNeeded; rating++ {
		username := fmt.Sprintf("Player_%d_3", rating)
		users = append(users, models.User{
			ID:       primitive.NewObjectID(),
			Username: username,
			Score:    rating,
		})
		extraAdded++
	}

	log.Printf("   Added %d extra users (3rd user for lower ratings)", extraAdded)
	log.Printf("   Total users to insert: %d", len(users))

	// Insert in batches with retry logic
	batchSize := 200 // Smaller batches for better reliability
	maxRetries := 3

	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		batch := users[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(users) + batchSize - 1) / batchSize

		var lastErr error
		for retry := 0; retry < maxRetries; retry++ {
			// Use a fresh context with longer timeout for each batch
			batchCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			_, err := collection.InsertMany(batchCtx, batch)
			cancel()

			if err == nil {
				log.Printf("   Inserted batch %d/%d (%d users)", batchNum, totalBatches, len(batch))
				lastErr = nil
				break
			}

			lastErr = err
			log.Printf("   ⚠️ Batch %d failed (attempt %d/%d): %v", batchNum, retry+1, maxRetries, err)

			// Wait before retry
			time.Sleep(time.Duration(2*(retry+1)) * time.Second)
		}

		if lastErr != nil {
			return 0, fmt.Errorf("failed to insert batch %d after %d retries: %w", batchNum, maxRetries, lastErr)
		}

		// Small delay between batches to avoid overwhelming the connection
		time.Sleep(100 * time.Millisecond)
	}

	// Re-initialize the leaderboard cache with a fresh context
	initCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := Initialize(initCtx); err != nil {
		return 0, fmt.Errorf("failed to initialize after seeding: %w", err)
	}

	log.Printf("✅ Successfully seeded %d users", len(users))
	return len(users), nil
}
