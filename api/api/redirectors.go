package api

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetRedirectorsHandler(c *gin.Context) {
	redirectors, err := data.GetRedirectors()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": redirectors})
	}
}

func CreateRedirectorHandler(c *gin.Context) {
	api_ip := os.Getenv("REACT_APP_NGINX_IP")
	api_port := os.Getenv("REACT_APP_NGINX_PORT")
	// redirector_port is static, we can use docker networking to set the host port at runtime
	redirector_port := os.Getenv("REDIRECTOR_PORT")

	newRedirectorID := uuid.New().String()
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	vForwardIP, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body["ForwardIP"])
	vForwardPort, _ := regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["ForwardPort"])
	vListenPort, _ := regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["ListenPort"])

	if !vForwardIP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ForwardIP"})
	} else if !vForwardPort {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ForwardPort"})
	} else if !vListenPort {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ListenPort"})
	} else {
		commandString := "docker save -o /app/payloads/redirector.tar patronc2/redirector"
		tmpl, err := template.ParseFiles("resources/redirector_install.sh.tmpl")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
			return
		}

		tplData := types.RedirectorTemplateData{
			LinkingKey:     newRedirectorID,
			ApiIP:          api_ip,
			ApiPort:        api_port,
			RedirectorPort: redirector_port,
			ExternalPort:   body["ListenPort"],
			ForwardIP:      body["ForwardIP"],
			ForwardPort:    body["ForwardPort"],
		}

		var scriptBuf bytes.Buffer
		if err := tmpl.Execute(&scriptBuf, tplData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render script"})
			return
		}

		script := scriptBuf.Bytes()

		logger.Logf(logger.Info, "Running command: %s", commandString)
		cmd := exec.Command("sh", "-c", commandString)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
		} else {
			data.CreateRedirector(newRedirectorID, body["Name"], body["Description"], body["ForwardIP"], body["ForwardPort"], body["ListenPort"])
			c.Header("Content-Disposition", "attachment; filename=redirector_install.sh")
			c.Data(http.StatusOK, "application/x-sh", script)
		}
	}
}

func RedirectorStatusHandler(c *gin.Context) {
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	linkingKey := body["linking_key"]
	err := data.SetRedirectorStatus(linkingKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}
