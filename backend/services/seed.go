// Package services contains the seeding logic for initial data.
package services

import (
	"context"
	"fmt"
	"log"

	"matiks-leaderboard/database"
	"matiks-leaderboard/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SeedDatabase(ctx context.Context) (int, error) {
	count, err := database.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	if count >= 10000 {
		log.Printf("📊 Database already has %d users, skipping seed", count)
		return 0, nil
	}

	needed := 10000 - int(count)
	log.Printf("🌱 Adding %d more users (current: %d, target: 10,000)...", needed, count)

	var users []interface{}

	// Track used usernames to ensure uniqueness
	usedUsernames := make(map[string]bool)

	// Fetch existing usernames to populate the map
	cursor, err := database.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		log.Printf("⚠️ Failed to fetch existing users: %v", err)
		// We continue, but duplicates might occur.
		// Ideally we should return error or handle it.
		// Since this is seeding, let's just log. Better to try-fail than stop?
		// Actually, let's return error to be safe as per user request to FIX it.
		return 0, err
	}
	defer cursor.Close(ctx)

	var existingUser struct {
		Username string `bson:"username"`
	}

	for cursor.Next(ctx) {
		if err := cursor.Decode(&existingUser); err == nil {
			usedUsernames[existingUser.Username] = true
		}
	}

	// Helper to generate unique username
	generateUsername := func(base string, index int) string {
		name := fmt.Sprintf("%s%d", base, index)
		if base == "User" {
			name = fmt.Sprintf("User%05d", index)
		}
		// Ensure absolute uniqueness by appending a counter if the generated name is already used
		originalName := name
		counter := 1
		for usedUsernames[name] {
			name = fmt.Sprintf("%s_%d", originalName, counter)
			counter++
		}
		usedUsernames[name] = true // Mark as used
		return name
	}

	// 1. Top Ratings (4900-5000): Max 2 users per score
	// 101 scores * 2 users = 202 users
	currentScore := 5000
	userCounter := 1

	for currentScore >= 4900 && len(users) < needed {
		// 2 users per score
		for i := 0; i < 2; i++ {
			username := generateUsername("TopPlayer", userCounter)
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Score:    currentScore,
			})
			addedUser(username, currentScore)
			userCounter++
		}
		currentScore--
	}

	// 2. Mid Ratings (2000-4899): 2 users per score
	// 2900 scores * 2 users = 5800 users
	for currentScore >= 2000 && len(users) < needed {
		for i := 0; i < 2; i++ {
			username := generateUsername("Player", userCounter)
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Score:    currentScore,
			})
			addedUser(username, currentScore)
			userCounter++
		}
		currentScore--
	}

	// 3. Lower Ratings (100-1999): 3 users per score (to fill up to 10000)
	// We need ~4000 more users.
	for currentScore >= 100 && len(users) < needed {
		// 3 users per score
		for i := 0; i < 3 && len(users) < needed; i++ {
			username := generateUsername("User", userCounter)
			users = append(users, models.User{
				ID:       primitive.NewObjectID(),
				Username: username,
				Score:    currentScore,
			})
			addedUser(username, currentScore)
			userCounter++
		}
		currentScore--
	}

	// Fill remaining if any
	for len(users) < needed {
		username := generateUsername("Newbie", userCounter)
		users = append(users, models.User{
			ID:       primitive.NewObjectID(),
			Username: username,
			Score:    100,
		})
		addedUser(username, 100)
		userCounter++
	}

	log.Printf("   Generated %d users", len(users))

	batchSize := 500
	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}
		_, err := database.Collection("users").InsertMany(ctx, users[i:end])
		if err != nil {
			return 0, err
		}
		log.Printf("   Inserted batch %d/%d", (i/batchSize)+1, (len(users)+batchSize-1)/batchSize)
	}

	if err := Initialize(ctx); err != nil {
		return 0, err
	}

	return len(users), nil
}

func addedUser(username string, score int) {
}
