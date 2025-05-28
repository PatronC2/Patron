package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	quic "github.com/quic-go/quic-go"
)

func main() {
	serverAddr := "172.17.206.245:9001"

	baseFreq := 30
	jitter := 20

	uuid := "test-uuid"

	for {
		err := sendConfigRequest(serverAddr, uuid)
		if err != nil {
			fmt.Printf("[-] Error: %v\n", err)
		}

		sleep := calculateCallbackSleep(baseFreq, jitter)
		fmt.Printf("[*] Sleeping for %d seconds\n", sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func sendConfigRequest(serverAddr, uuid string) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-patron"},
	}

	ctx := context.Background()

	session, err := quic.DialAddr(ctx, serverAddr, tlsConf, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer session.CloseWithError(0, "done")

	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		return fmt.Errorf("stream open failed: %w", err)
	}

	req := &patronobuf.Request{
		Type: patronobuf.RequestType_CONFIGURATION,
		Payload: &patronobuf.Request_Configuration{
			Configuration: &patronobuf.ConfigurationRequest{
				Uuid:              uuid,
				Username:          "testuser",
				Hostname:          "testhost",
				Ostype:            "windows",
				Arch:              "x64",
				Osbuild:           "10.0.19045",
				Cpus:              "4",
				Memory:            "16.0",
				Agentip:           "127.0.0.1",
				Serverip:          "127.0.0.1",
				Serverport:        "9000",
				Callbackfrequency: "30",
				Callbackjitter:    "20",
				Masterkey:         "MASTERKEY",
				NextcallbackUnix:  time.Now().Unix() + 30,
			},
		},
	}

	if err := common.WriteDelimited(stream, req); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(stream, resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	cfgResp := resp.GetConfigurationResponse()
	fmt.Printf("[+] Got config response: UUID=%s, CallbackFreq=%s\n", cfgResp.GetUuid(), cfgResp.GetCallbackfrequency())
	return nil
}

func calculateCallbackSleep(freq, jitter int) int {
	r := rand.Float64()
	jitterRange := float64(freq) * float64(jitter) / 100.0
	return freq - int(jitterRange) + int(r*2*jitterRange)
}
