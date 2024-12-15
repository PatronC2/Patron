package api

import (
	"net/http"
	"strconv"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
)

func ListTriggersForEventHandler(c *gin.Context) {
	eventIDstr := c.Param("eventID")

	eventID, err := strconv.Atoi(eventIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	triggers, err := data.ListTriggersByEvent(db, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list triggers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": triggers})
}

func CreateTriggerHandler(c *gin.Context) {
	var trigger types.Trigger
	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data"})
		return
	}

	triggerID, err := data.CreateTrigger(db, trigger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trigger"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"triggerID": triggerID})
}

func DeleteTriggerHandler(c *gin.Context) {
	triggerIDstr := c.Param("triggerID")

	triggerID, err := strconv.Atoi(triggerIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger ID"})
		return
	}
	if err := data.DeleteTrigger(db, triggerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trigger"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Trigger deleted successfully"})
}
