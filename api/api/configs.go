package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/gin-gonic/gin"
)

func GetLogLevelHandler(c *gin.Context) {
	app := c.Query("app")
	if app == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app query parameter"})
		return
	}

	level, err := data.GetLogLevel(app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch log level"})
		return
	}
	if level == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No log level found for app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"app": app, "log_level": level})
}

func SetLogLevelHandler(c *gin.Context) {
	app := c.Query("app")
	level := c.Query("log_level")

	if app == "" || level == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameters: app and log_level"})
		return
	}

	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warning": true,
		"error":   true,
	}
	if !validLevels[level] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "Invalid log_level value",
			"allowed_vals": []string{"debug", "info", "warning", "error"},
		})
		return
	}

	err := data.SetLogLevel(app, level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update log level"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Log level updated",
		"app":       app,
		"log_level": level,
	})
}

func GetLogFileSizeHandler(c *gin.Context) {
	app := c.Query("app")
	if app == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app query parameter"})
		return
	}

	sizeBytes, err := data.GetLogFileMaxSize(app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch log size"})
		return
	}

	sizeMB := sizeBytes / (1024 * 1024)

	c.JSON(http.StatusOK, gin.H{
		"app":            app,
		"size_bytes":     sizeBytes,
		"size_mb":        sizeMB,
		"human_readable": fmt.Sprintf("%d MB", sizeMB),
	})
}

type LogSizeUpdate struct {
	App  string `json:"app"`
	Size int64  `json:"size"`
	Unit string `json:"unit"`
}

func SetLogFileSizeHandler(c *gin.Context) {
	var req LogSizeUpdate

	if err := c.ShouldBindJSON(&req); err != nil || req.App == "" || req.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	// Convert size to bytes
	var sizeBytes int64
	switch strings.ToUpper(req.Unit) {
	case "MB":
		sizeBytes = req.Size * 1024 * 1024
	case "GB":
		sizeBytes = req.Size * 1024 * 1024 * 1024
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unit must be MB or GB"})
		return
	}

	err := data.SetLogFileMaxSize(req.App, sizeBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update log size"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Log file max size updated",
		"app":        req.App,
		"size_bytes": sizeBytes,
	})
}
