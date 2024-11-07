package api

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


func GetRedirectorsHandler(c *gin.Context) {
	redirectors, err := data.GetRedirectors()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error"})
	} else {
		c.JSON(http.StatusOK, gin.H{"redirectors": redirectors})
	}
}

func CreateRedirectorHandler(c *gin.Context) {
	api_ip			:= os.Getenv("REACT_APP_NGINX_IP")
	api_port		:= os.Getenv("REACT_APP_NGINX_PORT")
	redirector_port	:= os.Getenv("REDIRECTOR_PORT")

	newRedirectorID := uuid.New().String()
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	vForwardIP, _	:= regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body["ForwardIP"])	
	vForwardPort, _	:= regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["ForwardPort"])
	vListenIP, _		:= regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body["ListenIP"])
	vListenPort, _	:= regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["ListenPort"])

	if !vForwardIP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ForwardIP"})
	} else if !vForwardPort {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ForwardPort"})
	} else if !vListenIP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ListenIP"})
	} else if !vListenPort {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ListenPort"})
	} else {
		commandString := "docker save -o /app/payloads/redirector.tar patron-redirector"
		script := fmt.Sprintf(`
#!/bin/bash -xe

linking_key="%s"
tar_file="redirector.tar"

rm -f $tar_file

docker=$(which docker || echo "not found")
wget=$(which wget || echo "not found")

if [ -x "$docker" ]; then
    echo "Docker Check OK"
else
	# Remove conflicting Docker packages
	for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do
	sudo apt-get remove -y $pkg
	done

	# Install Docker
	sudo apt update
	sudo apt install -y ca-certificates curl
	sudo install -m 0755 -d /etc/apt/keyrings
	sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
	sudo chmod a+r /etc/apt/keyrings/docker.asc

	echo \
	"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
	$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
	sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

	sudo apt-get update
	sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

	sudo systemctl enable --now docker
	sudo docker --version
fi

if [ -x "$wget" ]; then
    echo "wget Check OK"
else
	sudo apt install -y wget
fi

docker image rm patron-redirector

wget --no-check-certificate https://%s:%s/files/$tar_file
docker load -i $tar_file
docker run -d \
	-p %s:%s \
	-e MAIN_SERVER_IP="%s" \
	-e MAIN_SERVER_PORT="%s" \
	-e FORWARDER_IP="%s" \
	-e FORWARDER_PORT="%s" \
	-v ./logs:/app/logs \
	patron-redirector`, newRedirectorID, api_ip, api_port, redirector_port, body["ListenPort"], body["ForwardIP"], body["ForwardPort"], body["ListenIP"], body["ListenPort"])

		logger.Logf(logger.Info, "Running command: %s", commandString)
		cmd := exec.Command("sh", "-c", commandString)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
		} else {
			data.CreateRedirector(newRedirectorID, body["name"], body["description"], body["ForwardIP"], body["ForwardPort"], body["ListenIP"], body["ListenPort"])
			c.Header("Content-Disposition", "attachment; filename=redirector_script.sh")
			c.Data(http.StatusOK, "application/x-sh", []byte(script))
		}
	}
}
