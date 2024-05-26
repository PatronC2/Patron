package main

import (
	"fmt"

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

    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}

func getGroupAgents(c *gin.Context) {
    // Get agent groups
    agentGroups, err := GroupAgentsByIp()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": agentGroups})
    return
}

func getGroupAgentsByIP(c *gin.Context) {
    // Get agents by IP
	ip := c.Param("ip")
    agents, err := AgentsByIp(ip)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}

func getOneAgentByUUID(c *gin.Context) {
    // Get agents by UUID
	uuid := c.Param("agt")
	fmt.Println("Trying to find agent", uuid)
    agents, err := FetchOne(uuid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
        return
    }
	// agents = [{69a124a8-6795-4481-a6a5-1a88e57a6e88 192.168.50.240 600 80    }]
    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}

func getAgentByUUID(c *gin.Context) {
    // Get agents by UUID
	uuid := c.Param("agt")
	fmt.Println("Trying to find agent", uuid)
    agents, err := Agent(uuid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}
