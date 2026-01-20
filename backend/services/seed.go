// Package services contains the seeding logic for initial data.
package services

import (
	"context"
	"log"

	"matiks-leaderboard/database"
	"matiks-leaderboard/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedDatabase populates the database with initial users if empty.
// Creates 10,000 users with controlled rank distribution:
// - Rank 1, 2, 3: One person each
// - Remaining: Two users per rank (same score)
func SeedDatabase(ctx context.Context) (int, error) {
	count, err := database.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	if count > 0 {
		log.Printf("📊 Database already has %d users, skipping seed", count)
		return 0, nil
	}

	log.Println("🌱 Generating 10,000+ users with controlled rank distribution...")

	prefixes := []string{
		"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta", "Iota", "Kappa",
		"Lambda", "Mu", "Nu", "Xi", "Omicron", "Pi", "Rho", "Sigma", "Tau", "Upsilon",
		"Phi", "Chi", "Psi", "Omega", "Phoenix", "Dragon", "Tiger", "Lion", "Eagle", "Falcon",
		"Shadow", "Storm", "Thunder", "Lightning", "Blaze", "Frost", "Fire", "Ice", "Wind", "Wave",
		"Ninja", "Samurai", "Knight", "Warrior", "Hunter", "Ranger", "Mage", "Wizard", "Sage", "Master",
		"Player", "Gamer", "Pro", "Elite", "Legend", "Hero", "King", "Queen", "Prince", "Duke",
		"Ace", "Star", "Nova", "Cosmic", "Cyber", "Tech", "Code", "Pixel", "Byte", "Vector",
		"Arun", "Sarthak", "Rahul", "Amit", "Vijay", "Raj", "Dev", "Max", "Alex", "Sam",
	}
	suffixes := []string{"", "Pro", "X", "Elite", "King", "One", "Max", "Star", "Legend", "Boss"}

	var users []interface{}
	usedUsernames := make(map[string]bool)

	addUser := func(baseName string, score int) {
		username := baseName
		attempt := 0
		for usedUsernames[username] {
			attempt++
			username = baseName + string(rune('0'+attempt))
		}
		usedUsernames[username] = true
		users = append(users, models.User{
			ID:       primitive.NewObjectID(),
			Username: username,
			Score:    score,
		})
	}

	// Top 3 ranks: single person each
	addUser("Champion", 5000)
	addUser("Legend", 4999)
	addUser("Master", 4998)

	// Remaining users: 2 per rank
	usernameIndex := 0
	currentScore := 4997

	for len(users) < 10000 && currentScore >= 100 {
		for i := 0; i < 2 && len(users) < 10000; i++ {
			prefixIdx := usernameIndex % len(prefixes)
			suffixIdx := (usernameIndex / len(prefixes)) % len(suffixes)
			numSuffix := usernameIndex / (len(prefixes) * len(suffixes))

			username := prefixes[prefixIdx] + suffixes[suffixIdx]
			if numSuffix > 0 {
				username += string(rune('0' + numSuffix))
			}

			addUser(username, currentScore)
			usernameIndex++
		}
		currentScore--
	}

	log.Printf("   Generated %d users", len(users))

	// Batch insert
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

	// Reload into cache
	if err := Initialize(ctx); err != nil {
		return 0, err
	}

	return len(users), nil
}
