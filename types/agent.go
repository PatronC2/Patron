package types

type configAgent struct {
	uuid              string
	callbackTo        string
	callbackFrequency int
	callbackJitter    float32
	// ...
}

//   type configListenerHttp struct {
// 	uri string
// 	// ...
//   }

type giveAgentCommand struct {
	updateAgentConfig configAgent
	// updateListenerConfig configListenerHttp
	commandType string // "execute, "upload", "download", etc., meterpreter style, or shell command like "whoami"
	command     string //
	binary      []byte // can be used with "upload" or "execute", etc.
	// ..
}

//sample := &giveAgentCommand{&configAgent{"1234", "192.20.20.12", 5, 4.5}, "shell", "whoami", nil }
