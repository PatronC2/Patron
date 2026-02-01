package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
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
	repoDir := os.Getenv("REPO_DIR")
	dockerHTTPSProxy := os.Getenv("DOCKER_HTTPS_PROXY")
	goVersion := os.Getenv("GOVERSION")

	configs, err := loadConfigurations("payloads.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
		return
	}

	newPayID := uuid.New().String()

	var req types.CreatePayloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateCreatePayload(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, exists := configs[req.Type]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}

	// Build output filename
	tag := strings.Split(newPayID, "-")
	concat := req.Name + "_" + tag[0] + config.FileSuffix

	// Dependencies
	var depCmd string
	if len(config.Dependencies) > 0 {
		parts := make([]string, 0, len(config.Dependencies))
		for _, dep := range config.Dependencies {
			parts = append(parts, fmt.Sprintf("go get %s", dep))
		}
		depCmd = strings.Join(parts, " && ") + " && "
	}

	// Build command
	commandString := fmt.Sprintf(
		"docker run --rm -v %s:/build -w /build -e HTTPS_PROXY=%s golang:%s sh -c '%s env %s go build %s \"-s -w "+
			"-X main.ServerIP=%s "+
			"-X main.ServerPort=%d "+
			"-X main.CallbackFrequency=%d "+
			"-X main.CallbackJitter=%d "+
			"-X main.RootCert=%s "+
			"-X main.LoggingEnabled=%t "+
			"-X main.TransportProtocol=%s\" "+
			"-o /build/payloads/%s /build/client/%s'",
		repoDir,
		dockerHTTPSProxy,
		goVersion,
		depCmd,
		config.Environment,
		config.Flags,
		req.ServerIP,
		req.ServerPort,
		req.CallbackFrequency,
		req.CallbackJitter,
		publickey,
		req.Logging,
		req.TransportProtocol,
		concat,
		config.CodePath,
	)

	logger.Logf(logger.Debug, "Running build command: %s", commandString)

	cmd := exec.Command("sh", "-c", commandString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Build failed",
			"details": err.Error(),
		})
		return
	}

	// Optional UPX compression
	if req.Compression == "upx" {
		upxCommand := fmt.Sprintf("upx --best --lzma /app/payloads/%s%s", concat, config.FileSuffix)
		logger.Logf(logger.Debug, "Running UPX command: %s", upxCommand)

		upxCmd := exec.Command("sh", "-c", upxCommand)
		upxCmd.Stdout = os.Stdout
		upxCmd.Stderr = os.Stderr
		if err := upxCmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "UPX compression failed",
				"details": err.Error(),
			})
			return
		}
	}

	p := types.Payload{
		Uuid:              newPayID,
		Name:              req.Name,
		Description:       req.Description,
		ServerIP:          req.ServerIP,
		Concat:            concat,
		TransportProtocol: req.TransportProtocol,
	}

	p.ServerPort = req.ServerPort
	p.CallbackFrequency = req.CallbackFrequency
	p.CallbackJitter = req.CallbackJitter

	if err := data.CreatePayload(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store payload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func validateCreatePayload(req types.CreatePayloadRequest) error {
	if req.ServerPort == 0 {
		return fmt.Errorf("serverport must be 1-65535")
	}
	if req.CallbackJitter > 100 {
		return fmt.Errorf("callbackjitter must be 0-100")
	}
	if req.Compression != "" && req.Compression != "upx" {
		return fmt.Errorf("compression must be empty or 'upx'")
	}
	return nil
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
	payloads, err := data.GetPayloads()
	if err != nil {
		logger.Logf(logger.Error, "Failed to get payloads from db: %v", err)
	}
	c.JSON(http.StatusOK, gin.H{"data": payloads})
}

func DeletePayloadHandler(c *gin.Context) {
	payloadIDStr := c.Param("payloadid")

	payloadID, err := strconv.ParseInt(payloadIDStr, 10, 64)
	if err != nil || payloadID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payloadid"})
		return
	}

	payloadConcat, err := data.GetPayloadConcat(payloadID)
	if err != nil {
		// If your DB returns sql.ErrNoRows, treat it as 404
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payload not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payload concat"})
		return
	}

	if err := data.DeletePayload(payloadID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payload not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payload"})
		return
	}

	if payloadConcat != "" {
		payloadPath := fmt.Sprintf("/app/payloads/%s", payloadConcat)
		if err := exec.Command("rm", "-f", payloadPath).Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payload from disk"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payload deleted successfully"})
}
