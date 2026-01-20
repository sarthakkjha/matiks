// Package services contains the seeding logic for initial data.
package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"matiks-leaderboard/database"
	"matiks-leaderboard/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Username prefixes for variety
var prefixes = []string{
	"Shadow", "Dragon", "Phoenix", "Storm", "Thunder", "Blaze", "Frost", "Night",
	"Star", "Moon", "Sun", "Fire", "Ice", "Dark", "Light", "Mystic", "Cyber",
	"Ninja", "Samurai", "Viking", "Knight", "Wizard", "Hunter", "Sniper", "Ghost",
	"Alpha", "Beta", "Omega", "Prime", "Elite", "Pro", "Master", "Legend",
	"Swift", "Rapid", "Turbo", "Hyper", "Ultra", "Mega", "Super", "Epic",
	"Ace", "King", "Queen", "Royal", "Crown", "Diamond", "Golden", "Silver",
	"Crimson", "Azure", "Cosmic", "Void", "Chaos", "Order", "Fury", "Rage",
}

var suffixes = []string{
	"X", "Z", "Pro", "HD", "XL", "Max", "Plus", "Prime", "Elite", "Master",
	"99", "007", "123", "360", "420", "777", "888", "1337", "2024", "3000",
}

// Special names to include (like Rahul)
var specialNames = []string{
	"Rahul", "Arjun", "Priya", "Neha", "Rohan", "Ananya", "Vikram", "Aisha",
	"Alex", "Jordan", "Sam", "Taylor", "Morgan", "Casey", "Riley", "Quinn",
	"Zara", "Leo", "Max", "Luna", "Nova", "Kai", "Ace", "Blaze",
}

// SeedDatabase creates 11,000 users with proper rating distribution.
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

	log.Println("🌱 Seeding 11,000 users with varied names...")

	var users []interface{}
	usedNames := make(map[string]bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Helper to generate unique username
	generateUniqueName := func(rating, index int) string {
		var name string

		// Different naming strategies based on rating tier
		switch {
		case rating >= 4500: // Top tier - Elite names
			prefix := prefixes[rng.Intn(len(prefixes))]
			name = fmt.Sprintf("%s_%d", prefix, rating)
		case rating >= 3500: // High tier - Prefix + suffix combo
			prefix := prefixes[rng.Intn(len(prefixes))]
			suffix := suffixes[rng.Intn(len(suffixes))]
			name = fmt.Sprintf("%s%s_%d", prefix, suffix, rating)
		case rating >= 2500: // Mid tier - Gaming style
			prefix := prefixes[rng.Intn(len(prefixes))]
			name = fmt.Sprintf("xX_%s_%d_Xx", prefix, rating)
		case rating >= 1500: // Lower-mid tier - Simple prefix + number
			prefix := prefixes[rng.Intn(len(prefixes))]
			name = fmt.Sprintf("%s%d_%d", prefix, rng.Intn(999), rating)
		default: // Low tier - Player style
			name = fmt.Sprintf("Player_%d_%d", rating, index)
		}

		// Ensure uniqueness
		originalName := name
		counter := 1
		for usedNames[name] {
			name = fmt.Sprintf("%s_%d", originalName, counter)
			counter++
		}
		usedNames[name] = true
		return name
	}

	// First: Add special names with high ratings (for demo purposes)
	for i, specialName := range specialNames {
		rating := 5000 - i // Rahul gets 5000, Arjun gets 4999, etc.
		if !usedNames[specialName] {
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: specialName,
				Score:    rating,
			})
			usedNames[specialName] = true
		}
	}
	log.Printf("   Added %d special names (including Rahul at #1)", len(specialNames))

	// Calculate remaining users needed
	remaining := 11000 - len(users)

	// Distribute across ratings 100-5000 with 2 users per rating
	// Some will get 3 for lower ratings to fill up
	ratingsNeeded := 4901 // 100 to 5000
	usersPerRating := 2
	extraUsersForLowRatings := remaining - (ratingsNeeded * usersPerRating)

	userIndex := 1
	for rating := 5000; rating >= 100 && len(users) < 11000; rating-- {
		// Skip ratings already used by special names
		count := usersPerRating
		if rating <= 100+extraUsersForLowRatings && extraUsersForLowRatings > 0 {
			count = 3
		}

		for i := 0; i < count && len(users) < 11000; i++ {
			username := generateUniqueName(rating, userIndex)
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Score:    rating,
			})
			userIndex++
		}
	}

	log.Printf("   Generated %d total users", len(users))

	// Insert in batches with retry logic
	batchSize := 200
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
			time.Sleep(time.Duration(2*(retry+1)) * time.Second)
		}

		if lastErr != nil {
			return 0, fmt.Errorf("failed to insert batch %d after %d retries: %w", batchNum, maxRetries, lastErr)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Re-initialize the leaderboard cache with a fresh context
	initCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := Initialize(initCtx); err != nil {
		return 0, fmt.Errorf("failed to initialize after seeding: %w", err)
	}

	log.Printf("✅ Successfully seeded %d users with varied names", len(users))
	return len(users), nil
}
