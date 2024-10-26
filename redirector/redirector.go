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

// Map to buffer data by UUID for each client
var buffer sync.Map

func main() {
	enableLogging := true
	logger.EnableLogging(enableLogging)

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

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept connection: %v\n", err)
			continue
		}
		go handleClientConnection(conn)
	}
}

func handleClientConnection(clientConn net.Conn) {
    defer clientConn.Close()
    clientAddr := clientConn.RemoteAddr().String()
    logger.Logf(logger.Info, "Accepted connection from %s\n", clientAddr)

    // Read data from client
    data := make([]byte, 4096)
    n, err := clientConn.Read(data)
    if err != nil {
        logger.Logf(logger.Warning, "Read error from client %s: %v\n", clientAddr, err)
        return
    }
    clientData := string(data[:n])

    // Parse UUID directly from clientData
    messageParts := strings.SplitN(clientData, ":", 3) // Adjust depending on data format
    if len(messageParts) < 3 {
        logger.Logf(logger.Warning, "Invalid data format from client %s", clientAddr)
        return
    }
    
    clientUUID := messageParts[0]
    if clientUUID == "" {
        logger.Logf(logger.Warning, "No UUID provided in client data from %s", clientAddr)
        return
    }
    buffer.Store(clientUUID, clientData)

    // Forward raw string data to main server
    err = forwardToMainServer(clientUUID, clientData)
    if err != nil {
        logger.Logf(logger.Warning, "Failed to forward to main server for client %s: %v\n", clientUUID, err)
        return
    }

    // Retrieve response data from buffer
    response, ok := buffer.Load(clientUUID)
    if !ok {
        logger.Logf(logger.Info, "No data available, sending empty response")
        gob.NewEncoder(clientConn).Encode([]byte{}) // Send empty response
    } else {
        logger.Logf(logger.Info, "Sending data to client: %v", response)
        gob.NewEncoder(clientConn).Encode(response)
        buffer.Delete(clientUUID) // Clear buffer after successful send
    }
}


// forwardToMainServer forwards client data to the main server and buffers the server's response
func forwardToMainServer(clientUUID string, clientData string) error {
    for {
        mainConn, err := connectToMainServer()
        if err != nil {
            logger.Logf(logger.Warning, "Main server unavailable, retrying in 5 seconds: %v\n", err)
            time.Sleep(5 * time.Second)
            continue
        }
        defer mainConn.Close()

        // Send raw client data as a string to the main server
        _, err = mainConn.Write([]byte(clientData))
        if err != nil {
            logger.Logf(logger.Warning, "Failed to send data to main server: %v\n", err)
            return err
        }

        // Decode the server's response into a byte array or a string if expected
        dec := gob.NewDecoder(mainConn)
        var serverResponse types.GiveAgentCommand
        if err := dec.Decode(&serverResponse); err != nil {
            logger.Logf(logger.Warning, "Failed to decode response from main server: %v\n", err)
            return err
        }

        logger.Logf(logger.Info, "Received server response: %v", serverResponse)
        buffer.Store(clientUUID, serverResponse)
        break
    }
    return nil
}


// connectToMainServer establishes a TLS connection to the main server
func connectToMainServer() (net.Conn, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	return tls.Dial("tcp", mainServerIP+":"+mainServerPort, tlsConfig)
}

// extractUUIDFromData extracts the UUID from the data message for use as the client identifier
func extractUUIDFromData(data []byte) string {
	// Split the data based on delimiter and validate
	parts := strings.Split(string(data), ":")
	if len(parts) > 0 && IsValidUUID(parts[0]) {
		return parts[0]
	}
	return ""
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
