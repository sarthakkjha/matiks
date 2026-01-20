// Package handlers contains HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"strconv"

	"matiks-leaderboard/services"

	"github.com/gin-gonic/gin"
)

func GetLeaderboard(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	response := services.GetLeaderboard(page, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func GetTopN(c *gin.Context) {
	n, _ := strconv.Atoi(c.Param("n"))
	if n < 1 {
		n = 10
	}
	if n > 100 {
		n = 100
	}

	entries := services.GetTopN(n)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"entries": entries, "count": len(entries)},
	})
}

func SearchUsers(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		prefix = c.Query("username")
	}
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "prefix is required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	users := services.SearchByPrefix(prefix, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"users": users, "count": len(users)},
	})
}

func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	user := services.GetUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Rating   int    `json:"rating"`
	Score    int    `json:"score"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	score := req.Rating
	if score == 0 {
		score = req.Score
	}
	if score == 0 {
		score = 100
	}

	user, err := services.CreateUser(c.Request.Context(), req.Username, score)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*services.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    gin.H{"user": user},
	})
}

type UpdateScoreRequest struct {
	Score  int `json:"score"`
	Rating int `json:"rating"`
}

func UpdateScore(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	score := req.Score
	if score == 0 {
		score = req.Rating
	}

	user, err := services.UpdateScore(c.Request.Context(), userID, score)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*services.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"user": user},
	})
}

type BulkUpdateRandomRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

func BulkUpdateRandom(c *gin.Context) {
	var req BulkUpdateRandomRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Count < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "count is required (min 1)",
		})
		return
	}

	result, err := services.BulkUpdateRandom(c.Request.Context(), req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

type BulkUpdateToValueRequest struct {
	Count  int `json:"count" binding:"required,min=1"`
	Rating int `json:"rating" binding:"required"`
}

func BulkUpdateToValue(c *gin.Context) {
	var req BulkUpdateToValueRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Count < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "count and rating are required",
		})
		return
	}

	result, err := services.BulkUpdateToValue(c.Request.Context(), req.Count, req.Rating)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*services.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    services.GetStats(),
	})
}
