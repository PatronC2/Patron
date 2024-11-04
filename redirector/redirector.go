package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/google/uuid"
)

// Forwarder settings
var (
	mainServerIP   = os.Getenv("MAIN_SERVER_IP")
	mainServerPort = os.Getenv("MAIN_SERVER_PORT")
	forwarderIP    = os.Getenv("FORWARDER_IP")
	forwarderPort  = os.Getenv("FORWARDER_PORT")
	certFile       = "certs/server.pem"
	keyFile        = "certs/server.key"
)

// Buffer to store client data by UUID
var buffer sync.Map

func main() {
	initializeLogging()
	cert, err := loadTLSCertificate(certFile, keyFile)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to load certificate: %v", err)
		return
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := startListener(forwarderIP, forwarderPort, tlsConfig)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to start listener: %v", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept connection: %v", err)
			continue
		}
		go handleClientConnection(conn)
	}
}

func initializeLogging() {
	logger.EnableLogging(true)
	if err := logger.SetLogFile("logs/forwarder.log"); err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
	}
}

func loadTLSCertificate(certFile, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certFile, keyFile)
}

func startListener(ip, port string, tlsConfig *tls.Config) (net.Listener, error) {
	addr := ip + ":" + port
	listener, err := tls.Listen("tcp", addr, tlsConfig)
	if err == nil {
		logger.Logf(logger.Info, "Forwarder listening on %s", addr)
	}
	return listener, err
}

func handleClientConnection(clientConn net.Conn) {
	defer clientConn.Close()

	for {
		mainConn, err := connectToMainServer()
		if err != nil {
			logger.Logf(logger.Warning, "Main server unavailable, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer mainConn.Close()

		clientAddr := clientConn.RemoteAddr().String()
		logger.Logf(logger.Info, "Accepted connection from %s", clientAddr)

		clientData, clientUUID, err := readInitialCallback(clientConn, clientAddr)
		if err != nil {
			return
		}
		buffer.Store(clientUUID, clientData)

		if err := forwardCallbackToMainServer(clientUUID, clientData, mainConn); err != nil {
			logger.Logf(logger.Warning, "Failed to forward to main server for client %s: %v", clientUUID, err)
			return
		}

		if err := sendCommandToClient(clientConn, clientUUID); err != nil {
			logger.Logf(logger.Warning, "Failed to send response to client %s: %v", clientUUID, err)
			return
		}

		if result, err := decodeClientResult(clientConn); err == nil {
			processCommandResult(clientUUID, result, mainConn)
		} else {
			logger.Logf(logger.Warning, "Failed to decode GiveServerResult from client %s: %v", clientAddr, err)
		}
		break
	}
}

func connectToMainServer() (*tls.Conn, error) {
	cert, err := loadTLSCertificate(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	return tls.Dial("tcp", mainServerIP+":"+mainServerPort, tlsConfig)
}

func readInitialCallback(clientConn net.Conn, clientAddr string) (string, string, error) {
	data := make([]byte, 4096)
	n, err := clientConn.Read(data)
	if err != nil {
		logger.Logf(logger.Warning, "Read error from client %s: %v", clientAddr, err)
		return "", "", err
	}

	clientData := string(data[:n])
	messageParts := strings.SplitN(clientData, ":", 3)
	if len(messageParts) < 3 {
		logger.Logf(logger.Warning, "Invalid data format from client %s", clientAddr)
		return "", "", fmt.Errorf("invalid data format")
	}

	clientUUID := messageParts[0]
	if clientUUID == "" {
		logger.Logf(logger.Warning, "No UUID provided in client data from %s", clientAddr)
		return "", "", fmt.Errorf("no UUID provided")
	}

	return clientData, clientUUID, nil
}

func forwardCallbackToMainServer(clientUUID, clientData string, mainConn *tls.Conn) error {
	if _, err := mainConn.Write([]byte(clientData)); err != nil {
		logger.Logf(logger.Warning, "Failed to send data to main server: %v", err)
		return err
	}

	var serverResponse types.GiveAgentCommand
	if err := gob.NewDecoder(mainConn).Decode(&serverResponse); err != nil {
		logger.Logf(logger.Warning, "Failed to decode response from main server: %v", err)
		return err
	}

	buffer.Store(clientUUID, serverResponse)
	return nil
}

func sendCommandToClient(clientConn net.Conn, clientUUID string) error {
	response, ok := buffer.Load(clientUUID)
	if !ok {
		logger.Logf(logger.Info, "No data available, sending empty response")
		return gob.NewEncoder(clientConn).Encode([]byte{})
	}

	logger.Logf(logger.Info, "Sending data to client: %v", response)
	if err := gob.NewEncoder(clientConn).Encode(response); err == nil {
		buffer.Delete(clientUUID)
	}
	return nil
}

func decodeClientResult(clientConn net.Conn) (types.GiveServerResult, error) {
	var result types.GiveServerResult
	err := gob.NewDecoder(clientConn).Decode(&result)
	return result, err
}

func processCommandResult(clientUUID string, result types.GiveServerResult, mainConn *tls.Conn) {
	logger.Logf(logger.Debug, "Sending command response: %v", result)
	if err := gob.NewEncoder(mainConn).Encode(result); err != nil {
		logger.Logf(logger.Error, "Error sending response: %v", err)
	}
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
