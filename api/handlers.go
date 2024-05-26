package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func getAgentsHandler(c *gin.Context) {
    // Get all agents
    agents, err := Agents()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"agents": agents})
    return
}

