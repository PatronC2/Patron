package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func loadConfigurations(filePath string) (types.PayloadConfigurations, error) {
	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var configs types.PayloadConfigurations
	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteValue, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func CreatePayloadHandler(c *gin.Context) {
	publickey := os.Getenv("PUBLIC_KEY")
	repo_dir := os.Getenv("REPO_DIR")

	configs, err := loadConfigurations("payloads.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
		return
	}

	newPayID := uuid.New().String()
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := validateBody(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := body["type"]
	config, exists := configs[key]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}

	tag := strings.Split(newPayID, "-")
	concat := body["name"] + "_" + tag[0] + config.FileSuffix

	dependencyCommands := ""
	for _, dep := range config.Dependencies {
		dependencyCommands += fmt.Sprintf("go get %s && ", dep)
	}

	commandString := fmt.Sprintf(
		"docker run --rm -v %s:/build -w /build -e HTTPS_PROXY=${DOCKER_HTTPS_PROXY} golang:1.22.3 sh -c '"+
			"%s env %s go build %s \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s -X main.LoggingEnabled=%s\" "+
			"-o /build/payloads/%s /build/client/%s'",
		repo_dir,
		dependencyCommands,
		config.Environment,
		config.Flags,
		body["serverip"],
		body["serverport"],
		body["callbackfrequency"],
		body["callbackjitter"],
		publickey,
		body["logging"],
		concat,
		config.CodePath,
	)

	fmt.Printf("Running build command: %s", commandString)
	cmd := exec.Command("sh", "-c", commandString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal Server Error", "details": err.Error()})
		return
	}

	if body["compression"] == "upx" {
		upxCommand := fmt.Sprintf("upx --best --lzma /app/payloads/%s%s", concat, config.FileSuffix)
		fmt.Printf("Running UPX command: %s", upxCommand)
		upxCmd := exec.Command("sh", "-c", upxCommand)
		upxCmd.Stdout = os.Stdout
		upxCmd.Stderr = os.Stderr
		err = upxCmd.Run()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "UPX compression failed", "details": err.Error()})
			return
		}
	}

	data.CreatePayload(newPayID, body["name"], body["description"], body["serverip"], body["serverport"], body["callbackfrequency"], body["callbackjitter"], concat)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func validateBody(body map[string]string) error {
	if net.ParseIP(body["serverip"]) == nil {
		return fmt.Errorf("invalid IP address")
	}

	port, err := strconv.Atoi(body["serverport"])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port")
	}

	callbackFrequency, err := strconv.Atoi(body["callbackfrequency"])
	if err != nil || callbackFrequency < 0 || callbackFrequency > 3600 {
		return fmt.Errorf("callbackfrequency must be a number between 0 and 3600")
	}

	callbackJitter, err := strconv.Atoi(body["callbackjitter"])
	if err != nil || callbackJitter < 1 || callbackJitter > 99 {
		return fmt.Errorf("callbackjitter must be a number between 1 and 99")
	}

	if strings.Contains(body["name"], " ") {
		return fmt.Errorf("name must not contain spaces")
	}

	if body["logging"] != "true" && body["logging"] != "false" {
		return fmt.Errorf("logging must be either 'true' or 'false'")
	}

	if body["compression"] != "none" && body["compression"] != "upx" {
		return fmt.Errorf("logging must be either 'none' or 'upx'")
	}

	return nil
}

func GetConfigurationsHandler(c *gin.Context) {
	configs, err := loadConfigurations("payloads.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
		return
	}
	c.JSON(http.StatusOK, configs)
}

func GetPayloadsHandler(c *gin.Context) {
	payloads := data.Payloads()
	c.JSON(http.StatusOK, gin.H{"data": payloads})
}

func DeletePayloadHandler(c *gin.Context) {
	payloadID := c.Param("payloadid")

	payloadConcat, err := data.GetPayloadConcat(payloadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payload name"})
		return
	}

	err = data.DeletePayload(payloadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payload"})
		return
	}

	payloadPath := fmt.Sprintf("/app/payloads/%s", payloadConcat)
	cmd := exec.Command("rm", "-f", payloadPath)

	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payload from disk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payload deleted successfully"})
}
