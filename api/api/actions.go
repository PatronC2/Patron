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

func ListActionsHandler(c *gin.Context) {
	actions, err := data.ListActions(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list actions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actions})
}

func GetActionHandler(c *gin.Context) {
	actionIDStr := c.Param("actionID")

	actionID, err := strconv.Atoi(actionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action ID"})
		return
	}

	action, err := data.GetActionByID(db, actionID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Action not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch action"})
		}
		return
	}

	c.JSON(http.StatusOK, action)
}

func CreateActionHandler(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve zip file"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open zip file"})
		return
	}
	defer src.Close()

	zipContent, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read zip file content"})
		return
	}

	action := types.Action{
		Name:        name,
		Description: description,
		File:        zipContent,
	}

	actionID, err := data.CreateAction(db, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"actionID": actionID})
}

func UpdateActionHandler(c *gin.Context) {
	actionIDStr := c.Param("actionID")

	actionID, err := strconv.Atoi(actionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action ID"})
		return
	}

	var action types.Action
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action data"})
		return
	}

	action.ActionID = actionID

	if err := data.UpdateAction(db, action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Action updated successfully"})
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
