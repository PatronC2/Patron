package api

import (
	"net"
	"net/http"
	"regexp"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAgentsHandler(c *gin.Context) {
	// Get all agents
	agents, err := data.Agents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

func GetGroupAgentsByIP(c *gin.Context) {
	// Get agents by IP
	ip := c.Param("ip")
	agents, err := data.AgentsByIp(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

func GetOneAgentByUUID(c *gin.Context) {
	// Get agents by UUID
	uuid := c.Param("agt")
	logger.Logf(logger.Debug, "Trying to find agent %v", uuid)
	agents, err := data.FetchOneAgent(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
		return
	}
	tags, err := data.GetAgentTags(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agents, "tags": tags})
}

func GetAgentCommandsByUUID(c *gin.Context) {
	// Get agents by UUID
	uuid := c.Param("agt")
	logger.Logf(logger.Debug, "Trying to find agent %v", uuid)
	agents, err := data.GetAgentCommands(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

func UpdateAgentHandler(c *gin.Context) {
	agentParam := c.Param("agt")

	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if net.ParseIP(body["serverip"]) == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid IP address"})
		return
	}

	vsvrPort := regexp.MustCompile(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`)
	if !vsvrPort.MatchString(body["serverport"]) {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid server port"})
		return
	}

	vfrequency := regexp.MustCompile(`^\d{1,5}$`)
	if !vfrequency.MatchString(body["callbackfreq"]) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Frequency, Max 99999"})
		return
	}

	vjitter := regexp.MustCompile(`^\d{1,5}$`)
	if !vjitter.MatchString(body["callbackjitter"]) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Jitter, Max 99999"})
		return
	}

	data.UpdateAgentConfig(agentParam, body["serverip"], body["serverport"], body["callbackfreq"], body["callbackjitter"])
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func SendCommandHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	newCmdID := uuid.New().String()

	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	command := body["command"]
	commandType := body["commandType"]
	data.SendAgentCommand(agentParam, "0", commandType, command, newCmdID)
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func KillAgentHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	newCmdID := uuid.New().String()
	data.SendAgentCommand(agentParam, "0", "kill", "Kill Agent", newCmdID)
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func GetKeylogHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	keylogs := data.Keylog(agentParam)
	c.JSON(http.StatusOK, gin.H{"data": keylogs})
}

func GetNoteHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	notes, err := data.GetAgentNotes(agentParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": notes})
	}
}

func PutNoteHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	notes := body["notes"]
	err := data.PutAgentNotes(agentParam, notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

func GetTagsHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	tags, err := data.GetAgentTags(agentParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"tags": tags})
	}
}

func PutTagsHandler(c *gin.Context) {
	var body struct {
		AgentUUIDs []string `json:"agents"`
		Key        string   `json:"key"`
		Value      string   `json:"value,omitempty"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Value == "" {
		body.Value = ""
	}

	for _, agentUUID := range body.AgentUUIDs {
		err := data.PutAgentTags(agentUUID, body.Key, body.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag for agent " + agentUUID})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags updated successfully for all agents"})
}

func DeleteTagHandler(c *gin.Context) {
	tagid := c.Param("tagid")
	err := data.DeleteTag(tagid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "deleted tag successfully"})
	}
}
