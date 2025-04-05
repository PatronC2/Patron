package api

import (
	"net/http"

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
