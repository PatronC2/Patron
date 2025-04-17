package types

import "time"

// Defining request and response types
type RequestType string
type ResponseType string

// Constants for request and response types
const (
	ConfigurationRequestType RequestType = "ConfigurationRequest"
	CommandRequestType       RequestType = "CommandRequest"
	CommandStatusRequestType RequestType = "CommandStatusRequest"
	KeysRequestType          RequestType = "KeysRequest"
	FileRequestType          RequestType = "FileRequest"
	FileToServerType         RequestType = "FileToServer"

	ConfigurationResponseType      ResponseType = "ConfigurationResponse"
	CommandResponseType            ResponseType = "CommandResponse"
	CommandStatusResponseType      ResponseType = "CommandStatusResponse"
	KeysResponseType               ResponseType = "KeysResponse"
	FileResponseType               ResponseType = "FileResponse"
	FileTransferStatusResponseType ResponseType = "FileTransferStatusResponse"
)

// General Request struct with typed payload
type Request struct {
	Type    RequestType
	Payload interface{}
}

// General Response struct with typed payload
type Response struct {
	Type    ResponseType
	Payload interface{}
}

// ConfigurationRequest is sent by agent to start a callback
type ConfigurationRequest struct {
	AgentID           string    `json:"uuid"`
	Username          string    `json:"username"`
	Hostname          string    `json:"hostname"`
	OSType            string    `json:"ostype"`
	OSArch            string    `json:"arch"`
	OSBuild           string    `json:"osbuild"`
	CPUS              string    `json:"cpus"`
	MEMORY            string    `json:"memory"`
	AgentIP           string    `json:"agentip"`
	ServerIP          string    `json:"serverip"`
	ServerPort        string    `json:"serverport"`
	CallbackFrequency string    `json:"callbackfrequency"`
	CallbackJitter    string    `json:"callbackjitter"`
	MasterKey         string    `json:"masterkey"`
	Status            string    `json:"status"`
	Tags              []Tag     `json:"tags"`
	NextCallback      time.Time `json:"nextcallback"`
}

// ConfigurationResponse is sent back to agent after a ConfigurationRequest
type ConfigurationResponse struct {
	AgentID           string `json:"uuid"`
	ServerIP          string `json:"serverip"`
	ServerPort        string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	CallbackJitter    string `json:"callbackjitter"`
}

// CommandRequest is sent by agent to check for commands
type CommandRequest struct {
	AgentID string `json:"uuid"`
}

// CommandResponse is sent back to the agent after a CommandRequest
type CommandResponse struct {
	AgentID     string `json:"uuid"`
	CommandType string `json:"commandtype"`
	CommandID   string `json:"commandid"`
	Command     string `json:"command"`
}

// CommandStatusRequest is sent by agent to tell the server the command outcome
type CommandStatusRequest struct {
	AgentID       string `json:"uuid"`
	CommandID     string `json:"commandid"`
	CommandResult string `json:"result"`
	CommandOutput string `json:"output"`
}

type CommandStatusResponse struct {
	AgentID string `json:"uuid"`
}

type KeysRequest struct {
	AgentID string `json:"uuid"`
	Keys    string `json:"keys"`
}

type KeysResponse struct {
	AgentID string `json:"uuid"`
}

type AgentCommands struct {
	Uuid        string `json:"uuid"`
	CommandType string `json:"commandtype"`
	Command     string `json:"command"`
	CommandUUID string `json:"commanduuid"`
	Output      string `json:"output"`
}

// FileRequest is sent by agent to check for files
type FileRequest struct {
	AgentID string `json:"uuid"`
}

// FileResponse is sent back to the agent after a FileRequest. Prepares the agent for a transfer
type FileResponse struct {
	FileID  string `json:"fileid"`
	AgentID string `json:"uuid"`
	Type    string `json:"transfertype"`
	Path    string `json:"filepath"`
	Chunk   []byte `json:"chunk"`
}

// This manages success messages if Type = download. Otherwise, send data to server
type FileToServer struct {
	FileID  string `json:"fileid"`
	AgentID string `json:"uuid"`
	Type    string `json:"transfertype"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	Chunk   []byte `json:"chunk"`
}

type FileTransferStatusResponse struct {
	FileID  string `json:"fileid"`
	AgentID string `json:"uuid"`
}
