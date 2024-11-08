package types

type ConfigAgent struct {
	Uuid              string `json:"uuid"`
	CallbackTo        string `json:"callbackto"`
	CallbackFrequency string `json:"callbackfrequency"`
	CallbackJitter    string `json:"callbackjitter"`
	AgentIP           string `json:"agentip"`
	Username          string `json:"username"`
	Hostname          string `json:"hostname"`
	Status            string `json:"status"`
	// ...
}

// type strictly for testing
type CreateAgentRequest struct {
	ServerIP         string `json:"serverip"`
	ServerPort       string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	Jitter           string `json:"jitter"`
	AgentIP          string `json:"agentip"`
	Username         string `json:"username"`
	Hostname         string `json:"hostname"`
}

type AgentIP struct {
	AgentIP           string `json:"agentip"`
	// ...
}

type GiveAgentCommand struct {
	UpdateAgentConfig ConfigAgent
	// updateListenerConfig configListenerHttp
	CommandType string // "execute, "upload", "download", etc., meterpreter style, or shell command like "whoami"
	Command     string //
	Binary      []byte // can be used with "upload" or "execute", etc.
	CommandUUID string
}

type GiveServerResult struct {
	Uuid        string
	Result      string
	Output      string
	CommandUUID string
}

type KeySend struct {
	Uuid string
	// Write  string
}

type KeyReceive struct {
	Uuid string `json:"uuid"`
	Keys string `json:"keys"`
}

//sample := &giveAgentCommand{&configAgent{"1234", "192.20.20.12", 5, 4.5}, "shell", "whoami", nil }

// // WEB Types

type Agent struct {
	Uuid        string `json:"uuid"`
	CommandType string `json:"commandtype"`
	Command     string `json:"command"`
	CommandUUID string `json:"commanduuid"`
	Output      string `json:"output"`
}

type Payload struct {
	Uuid              string `json:"uuid"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ServerIP          string `json:"serverip"`
	ServerPort        string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	CallbackJitter    string `json:"callbackjitter"`
	Concat            string `json:"concat"`
}

type BuildConfig struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CodePath    string `json:"code_path"`
	Flags       string `json:"flags"`
	Environment string `json:"environment"`
}

type PayloadConfigurations map[string]BuildConfig

