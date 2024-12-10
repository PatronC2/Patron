package api

import (
	"net/http"
	"strconv"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
)

func ListActionsHandler(c *gin.Context) {
	actions, err := data.ListActions(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list actions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actions})
}

func CreateActionHandler(c *gin.Context) {
	var action types.Action
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action data"})
		return
	}

	actionID, err := data.CreateAction(db, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"actionID": actionID})
}

func DeleteActionHandler(c *gin.Context) {
	actionIDStr := c.Param("actionID")

	actionID, err := strconv.Atoi(actionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action ID"})
		return
	}

	if err := data.DeleteAction(db, actionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Action deleted successfully"})
}
