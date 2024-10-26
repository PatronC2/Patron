package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/PatronC2/Patron/lib/logger"
)

var (
	mainServerIP   = os.Getenv("MAIN_SERVER_IP")
	mainServerPort = os.Getenv("MAIN_SERVER_PORT")
	forwarderIP    = os.Getenv("FORWARDER_IP")
	forwarderPort  = os.Getenv("FORWARDER_PORT")
	certFile       = "certs/server.pem"
	keyFile        = "certs/server.key"
)

func main() {
	// Enable or disable logging based on a condition
	enableLogging := true
	logger.EnableLogging(enableLogging)

	// Set the log file
	logFileName := "logs/redirector.log"
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
	}
	defer listener.Close()
	logger.Logf(logger.Info, "Forwarder listening on", forwarderIP+":"+forwarderPort)

	var buffer sync.Map

	// Accept incoming connections and handle each in a new goroutine
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept connection: %v\n", err)
			continue
		}

		go handleConnection(conn, &buffer)
	}
}

// handleConnection handles forwarding data to the main server with retries if it's down
func handleConnection(conn net.Conn, buffer *sync.Map) {
	clientAddr := conn.RemoteAddr().String()
	logger.Logf(logger.Info, "Accepted connection from %s\n", clientAddr)

	for {
		// Connect to the main server
		mainConn, err := connectToMainServer()
		if err != nil {
			logger.Logf(logger.Warning, "Patron is unavailable. Retrying in 5 seconds. %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer mainConn.Close()

		// Forward data to main server with buffering
		go forwardWithBuffer(conn, mainConn, buffer, clientAddr)
		break
	}
}

// connectToMainServer establishes a TLS connection to the main server
func connectToMainServer() (net.Conn, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load certificate: %w\n", err)
    }

    tlsConfig := &tls.Config{
        Certificates:       []tls.Certificate{cert},
        InsecureSkipVerify: true,
    }
    
    return tls.Dial("tcp", mainServerIP+":"+mainServerPort, tlsConfig)
}

// Updated forwardWithBuffer to handle read errors and retry logic
func forwardWithBuffer(clientConn, mainConn net.Conn, buffer *sync.Map, clientAddr string) {
    for {
        data := make([]byte, 4096)
        n, err := clientConn.Read(data)
        if err != nil {
            if err == io.EOF {
                logger.Logf(logger.Info, "Client %s disconnected\n", clientAddr)
                break
            }
            logger.Logf(logger.Warning, "Read error from client %s: %v\n", clientAddr, err)
            buffer.Store(clientAddr, data[:n])
            retrySendBuffer(mainConn, buffer, clientAddr)
            return
        }

        // Attempt to forward data, buffer if it fails
        _, writeErr := mainConn.Write(data[:n])
        if writeErr != nil {
            logger.Logf(logger.Warning, "Main server unavailable, buffering data for client %s\n", clientAddr)
            buffer.Store(clientAddr, data[:n])
            retrySendBuffer(mainConn, buffer, clientAddr)
            return
        }
    }
}

// Updated retrySendBuffer to reattempt sending buffered data without deferring close
func retrySendBuffer(mainConn net.Conn, buffer *sync.Map, clientAddr string) {
    for {
        time.Sleep(5 * time.Second)
        var err error
        if mainConn == nil {
            mainConn, err = connectToMainServer()
            if err != nil {
                logger.Logf(logger.Warning, "Main server still unavailable. Retrying. %v\n", err)
                continue
            }
        }

		if val, ok := buffer.Load(clientAddr); ok {
			// Ensure val is indeed a []byte
			if data, ok := val.([]byte); ok {
				_, err := mainConn.Write(data)
				if err == nil {
					buffer.Delete(clientAddr)
					logger.Logf(logger.Info, "Buffered data sent to main server for client %s\n", clientAddr)
					logger.Logf(logger.Debug, "Buffered data: %q\n", data) // Safe printing
					break
				} else {
					logger.Logf(logger.Warning, "Failed to send buffered data: %v", err)
					mainConn.Close()
					mainConn = nil
				}
			} else {
				logger.Logf(logger.Warning, "Invalid type for buffered data for client %s", clientAddr)
			}
		}		
    }
}

