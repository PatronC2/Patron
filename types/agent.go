package types

type ConfigAgent struct {
	Uuid              string
	CallbackTo        string
	CallbackFrequency int
	CallbackJitter    float32
	// ...
}

//   type configListenerHttp struct {
// 	uri string
// 	// ...
//   }

type GiveAgentCommand struct {
	UpdateAgentConfig ConfigAgent
	// updateListenerConfig configListenerHttp
	CommandType string // "execute, "upload", "download", etc., meterpreter style, or shell command like "whoami"
	Command     string //
	Binary      []byte // can be used with "upload" or "execute", etc.
	// ..
}

type GiveServerResult struct {
	Uuid   string
	Result string
	Output string
}

//sample := &giveAgentCommand{&configAgent{"1234", "192.20.20.12", 5, 4.5}, "shell", "whoami", nil }