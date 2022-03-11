package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func handleconn(connection net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		// texted := strings.Split(text, "\n")
		// send := texted[0] + "end"
		// fmt.Print()
		fmt.Fprintf(connection, text)
		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Print("->: " + message)

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
