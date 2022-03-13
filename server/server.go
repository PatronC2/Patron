package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// TODO add logging
// client connect, callback

func handleconn(connection net.Conn) {
	for {
		encoder := gob.NewEncoder(connection)
		instruct := &agent.giveAgentCommand{&agent.configAgent{"1234", "192.20.20.12", 5, 4.5}, "shell", "whoami", nil}
		encoder.Encode(instruct)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprint(connection, text)
		message := make([]byte, 4096)
		connection.Read(message)
		out := string(message)
		fmt.Println(out)
		// message, _ := bufio.NewReader(connection)

		// message, _ := bufio.NewReader(connection).ReadString('\n')
		// re := regexp.MustCompile(`~w`)
		// message = re.ReplaceAllString(message, "\n")
		// fmt.Print("->: " + message)

		if strings.TrimSpace(string(message)) == "STOP" {
			fmt.Println("Exiting TCP server!")
			break
		}
	}
	connection.Close()
}

func main() {
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.
			Fatalln(err)
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
