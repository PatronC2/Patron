package types

type Payload struct {
	PayloadID         int64  `json:"payloadid"`
	Uuid              string `json:"uuid"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ServerIP          string `json:"serverip" binding:"required"`
	ServerPort        uint16 `json:"serverport" binding:"required"`
	CallbackFrequency uint32 `json:"callbackfrequency" binding:"required"`
	CallbackJitter    uint8  `json:"callbackjitter" binding:"required"`
	Concat            string `json:"concat"`
	TransportProtocol string `json:"transportprotocol"`
}

type BuildConfig struct {
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	OS           string   `json:"os"`
	CodePath     string   `json:"code_path"`
	Flags        string   `json:"flags"`
	Environment  string   `json:"environment"`
	FileSuffix   string   `json:"file_suffix"`
	Dependencies []string `json:"dependencies"`
}

type CreatePayloadRequest struct {
	Type              string `json:"type" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	ServerIP          string `json:"serverip" binding:"required"`
	ServerPort        uint16 `json:"serverport" binding:"required"`
	CallbackFrequency uint32 `json:"callbackfrequency" binding:"required"`
	CallbackJitter    uint8  `json:"callbackjitter" binding:"required"`

	TransportProtocol string `json:"transportprotocol" binding:"required"`
	Logging           bool   `json:"logging"`
	Compression       string `json:"compression"`
}

type PayloadConfigurations map[string]BuildConfig
