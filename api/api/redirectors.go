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
		script := fmt.Sprintf(`#!/bin/bash -xe

linking_key="%s"
api_ip="%s"
api_port="%s"
# This is the port set when the app was compiled. Do not change.
redirector_port=%s
# This port can be freely changed to change the port that the redirector listens on.
external_redirector_port=%s
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

daemon_file="/etc/docker/daemon.json"
desired_config='{"ipv6": true, "fixed-cidr-v6": "2001:db8:1::/64"}'

if [ -f "$daemon_file" ]; then
    current_config=$(jq -c . < "$daemon_file" 2>/dev/null || echo "{}")
    normalized_desired_config=$(echo "$desired_config" | jq -c .)

    if [ "$current_config" != "$normalized_desired_config" ]; then
        echo "Updating Docker daemon.json..."
        echo "$normalized_desired_config" > "$daemon_file"
        systemctl restart docker
    else
        echo "Docker daemon.json already configured as desired. Skipping update and restart."
    fi
else
    echo "Creating Docker daemon.json with desired configuration..."
    echo "$desired_config" > "$daemon_file"
    systemctl restart docker
fi

if [ -x "$wget" ]; then
    echo "wget Check OK"
else
	sudo apt install -y wget
fi

wget --no-check-certificate https://$api_ip:$api_port/fileserver/$tar_file
docker load -i $tar_file
docker run -d \
	-p $external_redirector_port:$redirector_port \
	-e MAIN_SERVER_IP="%s" \
	-e MAIN_SERVER_PORT="%s" \
	-e FORWARDER_PORT="$redirector_port" \
	-e LINKING_KEY="$linking_key" \
	-e API_IP="$api_ip" \
	-e API_PORT="$api_port" \
	-v ./logs:/app/logs \
	patronc2/redirector
`, newRedirectorID, api_ip, api_port, redirector_port, body["ListenPort"], body["ForwardIP"], body["ForwardPort"])

		logger.Logf(logger.Info, "Running command: %s", commandString)
		cmd := exec.Command("sh", "-c", commandString)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
		} else {
			data.CreateRedirector(newRedirectorID, body["Name"], body["Description"], body["ForwardIP"], body["ForwardPort"], body["ListenPort"])
			c.Header("Content-Disposition", "attachment; filename=redirector_install.sh")
			c.Data(http.StatusOK, "application/x-sh", []byte(script))
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
