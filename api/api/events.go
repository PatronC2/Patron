package api

import (
	"database/sql"
	"io"
	"net/http"
	"strconv"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
)

func ListEventsHandler(c *gin.Context) {
	events, err := data.ListEvents(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

func GetEventHandler(c *gin.Context) {
	eventIDStr := c.Param("eventID")

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := data.GetEventByID(db, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event"})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

func CreateEventHandler(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	schedule := c.PostForm("schedule")

	file, err := c.FormFile("script")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve script file"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open script file"})
		return
	}
	defer src.Close()

	scriptContent, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read script file content"})
		return
	}

	event := types.Event{
		Name:        name,
		Description: description,
		Script:      scriptContent,
		Schedule:    schedule,
	}

	eventID, err := data.CreateEvent(db, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"eventID": eventID})
}

func UpdateEventHandler(c *gin.Context) {
	eventIDStr := c.Param("eventID")

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event types.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event data"})
		return
	}

	event.EventID = eventID

	if err := data.UpdateEvent(db, event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Event updated successfully"})
}

func DeleteEventHandler(c *gin.Context) {
	eventIDStr := c.Param("eventID")

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := data.DeleteEvent(db, eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Event deleted successfully"})
}
