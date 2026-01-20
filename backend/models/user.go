// Package models defines the data structures used throughout the application.
// These models represent the core entities in the leaderboard system.
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents a player in the leaderboard system.
// Stored in MongoDB with username and score fields.
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Score    int                `bson:"score" json:"score"`
}

// UserResponse is the JSON response format for API endpoints.
// Includes computed rank from the ranking engine.
type UserResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank,omitempty"`
}

// LeaderboardEntry represents a single entry in the leaderboard.
// Includes rank computed from the snapshot manager.
type LeaderboardEntry struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

// LeaderboardResponse is the paginated response for leaderboard queries.
type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	TotalUsers int                `json:"totalUsers"`
	TotalPages int                `json:"totalPages"`
	Page       int                `json:"page"`
}

// BulkUpdateResult contains the results of a bulk update operation.
type BulkUpdateResult struct {
	Updated       int     `json:"updated"`
	DurationMs    int64   `json:"durationMs"`
	UpdatesPerSec float64 `json:"updatesPerSec"`
}
