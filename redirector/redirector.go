package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	quic "github.com/quic-go/quic-go"
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

type netConnWrapper struct {
	net.Conn
}

type TransportType int

const (
	TransportTCP TransportType = iota
	TransportQUIC
)

func (n netConnWrapper) Close() error { return n.Conn.Close() }

func main() {
	enableLogging := true
	logger.EnableLogging(enableLogging)

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
	go startQUICListener("0.0.0.0:"+forwarderPort, tlsConfig)

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
		go handleClientConnection(conn, TransportTCP)
	}
}

func startQUICListener(address string, baseTLSConfig *tls.Config) {
	quicConfig := &quic.Config{}

	tlsConf := baseTLSConfig.Clone()
	tlsConf.NextProtos = []string{"quic-patron"}

	listener, err := quic.ListenAddr(address, tlsConf, quicConfig)
	if err != nil {
		logger.Logf(logger.Warning, "Failed to start QUIC listener on %s: %v", address, err)
		return
	}
	logger.Logf(logger.Info, "QUIC Forwarder listening on %s", address)

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			logger.Logf(logger.Warning, "Failed to accept QUIC connection: %v", err)
			continue
		}
		go func(conn quic.Connection) {
			for {
				stream, err := conn.AcceptStream(context.Background())
				if err != nil {
					logger.Logf(logger.Warning, "Failed to accept QUIC stream: %v", err)
					return
				}
				go handleClientConnection(stream, TransportQUIC)
			}
		}(conn)
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

func handleClientConnection(clientConn types.CommonStream, transport TransportType) {
	defer clientConn.Close()

	var (
		mainConn types.CommonStream
		err      error
	)

	switch transport {
	case TransportQUIC:
		mainConn, err = connectToMainServerQUIC()
	case TransportTCP:
		mainConn, err = connectToMainServer()
	default:
		logger.Logf(logger.Error, "Unknown transport type for forwarding")
		return
	}

	if err != nil {
		logger.Logf(logger.Warning, "Failed to connect to main server: %v", err)
		return
	}
	defer mainConn.Close()

	for {
		var request patronobuf.Request
		if err := common.ReadDelimited(clientConn, &request); err != nil {
			logger.Logf(logger.Warning, "Failed to read from client: %v", err)
			return
		}

		if err := common.WriteDelimited(mainConn, &request); err != nil {
			logger.Logf(logger.Warning, "Failed to forward to server: %v", err)
			return
		}

		var response patronobuf.Response
		if err := common.ReadDelimited(mainConn, &response); err != nil {
			logger.Logf(logger.Warning, "Failed to read from server: %v", err)
			return
		}

		if err := common.WriteDelimited(clientConn, &response); err != nil {
			logger.Logf(logger.Warning, "Failed to send response to client: %v", err)
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

func connectToMainServerQUIC() (quic.Stream, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-patron"},
	}
	session, err := quic.DialAddr(context.Background(), mainServerIP+":"+mainServerPort, tlsConf, &quic.Config{})
	if err != nil {
		return nil, err
	}
	return session.OpenStreamSync(context.Background())
}
