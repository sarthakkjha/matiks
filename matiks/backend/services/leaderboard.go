// Package services contains the business logic for the leaderboard.
// Coordinates database operations, caching, and ranking engine updates.
// Implements debounced rebuilds for high-throughput update handling.
package services

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"matiks-leaderboard/cache"
	"matiks-leaderboard/database"
	"matiks-leaderboard/engine"
	"matiks-leaderboard/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Configuration for high-throughput updates.
// SCALABILITY STRATEGY:
// Instead of rebuilding the leaderboard on every single update (expensive O(N log N)),
// we "debounce" the rebuilds. We wait for a quiet period (100ms) or max delay (500ms)
// and aggregate all pending updates into a SINGLE rebuild operation.
//
// Result: 100 updates/sec -> 1 rebuild (instead of 100 rebuilds).
const (
	RebuildDelayMS    = 100 // Wait 100ms for more updates to arrive
	MaxRebuildDelayMS = 500 // Force rebuild if we've waited this long (prevent staleness)
)

// Stats tracks update statistics for monitoring.
type Stats struct {
	mu                   sync.RWMutex
	TotalUpdates         int64
	RebuildsTriggered    int64
	AvgUpdatesPerRebuild float64
}

var (
	stats          = &Stats{}
	pendingUpdates int64
	rebuildTimer   *time.Timer
	lastRebuild    time.Time
	rebuildMu      sync.Mutex
)

// Initialize loads all users from MongoDB into cache and builds the snapshot.
// Called once at startup.
func Initialize(ctx context.Context) error {
	cursor, err := database.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	cache.Global.Clear()
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		cache.Global.Set(user.ID.Hex(), cache.Entry{
			Username: user.Username,
			Score:    user.Score,
		})
	}

	ForceRebuild()
	log.Printf("✅ Loaded %d users into cache", cache.Global.Size())
	return nil
}

// GetLeaderboard returns paginated leaderboard data.
func GetLeaderboard(page, limit int) *models.LeaderboardResponse {
	entries, total := engine.Global.GetLeaderboard(page, limit)

	result := make([]models.LeaderboardEntry, len(entries))
	for i, e := range entries {
		result[i] = models.LeaderboardEntry{
			UserID:   e.UserID,
			Username: e.Username,
			Rating:   e.Score,
			Rank:     e.Rank,
		}
	}

	return &models.LeaderboardResponse{
		Entries:    result,
		TotalUsers: total,
		TotalPages: (total + limit - 1) / limit,
		Page:       page,
	}
}

// GetTopN returns the top N users.
func GetTopN(n int) []models.LeaderboardEntry {
	entries := engine.Global.GetTop(n)

	result := make([]models.LeaderboardEntry, len(entries))
	for i, e := range entries {
		result[i] = models.LeaderboardEntry{
			UserID:   e.UserID,
			Username: e.Username,
			Rating:   e.Score,
			Rank:     e.Rank,
		}
	}
	return result
}

// SearchByPrefix searches users by username prefix.
// Returns results with their current rank.
func SearchByPrefix(prefix string, limit int) []models.UserResponse {
	results := cache.Global.SearchByPrefix(prefix, limit)

	users := make([]models.UserResponse, len(results))
	for i, r := range results {
		users[i] = models.UserResponse{
			UserID:   r.UserID,
			Username: r.Username,
			Rating:   r.Score,
			Rank:     engine.Global.GetRank(r.UserID),
		}
	}
	return users
}

// GetUserByID retrieves a user by ID with their rank.
func GetUserByID(userID string) *models.UserResponse {
	entry, ok := cache.Global.Get(userID)
	if !ok {
		return nil
	}

	return &models.UserResponse{
		UserID:   userID,
		Username: entry.Username,
		Rating:   entry.Score,
		Rank:     engine.Global.GetRank(userID),
	}
}

// CreateUser creates a new user in the database.
func CreateUser(ctx context.Context, username string, score int) (*models.UserResponse, error) {
	if score < 100 || score > 5000 {
		return nil, &ValidationError{"Score must be between 100 and 5000"}
	}

	user := models.User{Username: username, Score: score}
	result, err := database.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	userID := result.InsertedID.(primitive.ObjectID).Hex()
	cache.Global.Set(userID, cache.Entry{Username: username, Score: score})
	scheduleRebuild()

	return &models.UserResponse{
		UserID:   userID,
		Username: username,
		Rating:   score,
	}, nil
}

// UpdateScore updates a user's score.
// Cache is updated immediately; snapshot rebuild is debounced.
func UpdateScore(ctx context.Context, userID string, newScore int) (*models.UserResponse, error) {
	if newScore < 100 || newScore > 5000 {
		return nil, &ValidationError{"Score must be between 100 and 5000"}
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = database.Collection("users").FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"score": newScore}},
	).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Update cache immediately (O(1))
	cache.Global.Set(userID, cache.Entry{Username: user.Username, Score: newScore})
	// Schedule debounced rebuild
	scheduleRebuild()

	return &models.UserResponse{
		UserID:   userID,
		Username: user.Username,
		Rating:   newScore,
		Rank:     engine.Global.GetRank(userID),
	}, nil
}

// BulkUpdateRandom updates 'count' random users with random scores.
// Returns performance metrics.
func BulkUpdateRandom(ctx context.Context, count int) (*models.BulkUpdateResult, error) {
	start := time.Now()

	allUsers := cache.Global.GetAllWithIDs()
	userIDs := make([]string, 0, len(allUsers))
	for id := range allUsers {
		userIDs = append(userIDs, id)
	}

	if count > len(userIDs) {
		count = len(userIDs)
	}

	// Shuffle for randomness
	rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})
	userIDs = userIDs[:count]

	updated := 0
	for _, id := range userIDs {
		newScore := rand.Intn(4901) + 100
		objID, _ := primitive.ObjectIDFromHex(id)

		_, err := database.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{"score": newScore}},
		)
		if err == nil {
			entry, _ := cache.Global.Get(id)
			cache.Global.Set(id, cache.Entry{Username: entry.Username, Score: newScore})
			updated++
		}
	}

	ForceRebuild()
	duration := time.Since(start)

	return &models.BulkUpdateResult{
		Updated:       updated,
		DurationMs:    duration.Milliseconds(),
		UpdatesPerSec: float64(updated) / duration.Seconds(),
	}, nil
}

// BulkUpdateToValue updates 'count' random users to a specific score.
func BulkUpdateToValue(ctx context.Context, count, targetScore int) (*models.BulkUpdateResult, error) {
	if targetScore < 100 || targetScore > 5000 {
		return nil, &ValidationError{"Score must be between 100 and 5000"}
	}

	start := time.Now()

	allUsers := cache.Global.GetAllWithIDs()
	userIDs := make([]string, 0, len(allUsers))
	for id := range allUsers {
		userIDs = append(userIDs, id)
	}

	if count > len(userIDs) {
		count = len(userIDs)
	}

	rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})
	userIDs = userIDs[:count]

	updated := 0
	for _, id := range userIDs {
		objID, _ := primitive.ObjectIDFromHex(id)

		_, err := database.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{"score": targetScore}},
		)
		if err == nil {
			entry, _ := cache.Global.Get(id)
			cache.Global.Set(id, cache.Entry{Username: entry.Username, Score: targetScore})
			updated++
		}
	}

	ForceRebuild()
	duration := time.Since(start)

	return &models.BulkUpdateResult{
		Updated:       updated,
		DurationMs:    duration.Milliseconds(),
		UpdatesPerSec: float64(updated) / duration.Seconds(),
	}, nil
}

// GetStats returns service statistics for monitoring.
func GetStats() map[string]interface{} {
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	return map[string]interface{}{
		"totalUsers":           cache.Global.Size(),
		"pendingUpdates":       pendingUpdates,
		"totalUpdates":         stats.TotalUpdates,
		"rebuildsTriggered":    stats.RebuildsTriggered,
		"avgUpdatesPerRebuild": stats.AvgUpdatesPerRebuild,
	}
}

// scheduleRebuild implements debounced rebuilding for high-throughput.
func scheduleRebuild() {
	rebuildMu.Lock()
	defer rebuildMu.Unlock()

	pendingUpdates++
	stats.mu.Lock()
	stats.TotalUpdates++
	stats.mu.Unlock()

	if time.Since(lastRebuild) >= MaxRebuildDelayMS*time.Millisecond && pendingUpdates > 0 {
		executeRebuild()
		return
	}

	if rebuildTimer != nil {
		rebuildTimer.Stop()
	}
	rebuildTimer = time.AfterFunc(RebuildDelayMS*time.Millisecond, func() {
		rebuildMu.Lock()
		defer rebuildMu.Unlock()
		executeRebuild()
	})
}

// executeRebuild performs the actual snapshot rebuild.
func executeRebuild() {
	count := pendingUpdates
	pendingUpdates = 0
	lastRebuild = time.Now()

	stats.mu.Lock()
	stats.RebuildsTriggered++
	if stats.RebuildsTriggered > 0 {
		stats.AvgUpdatesPerRebuild = float64(stats.TotalUpdates) / float64(stats.RebuildsTriggered)
	}
	stats.mu.Unlock()

	engine.Global.Rebuild(cache.Global.GetAllWithIDs())
	log.Printf("🔄 Snapshot rebuilt (batched %d updates)", count)
}

// ForceRebuild immediately rebuilds the snapshot.
func ForceRebuild() {
	rebuildMu.Lock()
	defer rebuildMu.Unlock()

	if rebuildTimer != nil {
		rebuildTimer.Stop()
	}
	pendingUpdates = 0
	lastRebuild = time.Now()
	engine.Global.Rebuild(cache.Global.GetAllWithIDs())
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
