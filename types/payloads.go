package types

type Payload struct {
	PayloadID         string `json:"payloadid"`
	Uuid              string `json:"uuid"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ServerIP          string `json:"serverip"`
	ServerPort        string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	CallbackJitter    string `json:"callbackjitter"`
	Concat            string `json:"concat"`
	TransportProtocol string `json:"transportprotocol"`
}

type BuildConfig struct {
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	CodePath     string   `json:"code_path"`
	Flags        string   `json:"flags"`
	Environment  string   `json:"environment"`
	FileSuffix   string   `json:"file_suffix"`
	Dependencies []string `json:"dependencies"`
}

type PayloadConfigurations map[string]BuildConfig
