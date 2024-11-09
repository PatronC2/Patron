package main

import (
	"crypto/tls"
	"encoding/gob"
	"log"
	"net"
	"os"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

func validateOrCreateAgent(c types.ConfigurationRequest) (types.ConfigurationResponse, bool) {
	fetch, err := data.FetchOneAgent(c.AgentID)
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return types.ConfigurationResponse{}, false
	}

	if fetch.AgentID == "" && c.MasterKey == "MASTERKEY" {
		data.CreateAgent(c.AgentID, c.ServerIP, c.ServerPort, c.CallbackFrequency, c.CallbackJitter, c.AgentIP, c.Username, c.Hostname)
		data.CreateKeys(c.AgentID)
		fetch, err = data.FetchOneAgent(c.AgentID)
		if err != nil {
			logger.Logf(logger.Warning, "Couldn't fetch agent after creation: %v\n", err)
			return types.ConfigurationResponse{}, false
		}
	}

	response := types.ConfigurationResponse{
		ServerIP:         fetch.ServerIP,
		ServerPort:       fetch.ServerPort,
		CallbackFrequency: fetch.CallbackFrequency,
		CallbackJitter:   fetch.CallbackJitter,
	}

	return response, fetch.AgentID == c.AgentID
}

type ConfigurationHandler struct{}

func (h *ConfigurationHandler) Handle(request types.Request, conn net.Conn) types.Response {
    configReq, ok := request.Payload.(types.ConfigurationRequest)
    if !ok {
        return types.Response{
            Type:    types.ConfigurationResponseType,
            Payload: types.ConfigurationResponse{},
        }
    }

    configResponse, success := validateOrCreateAgent(configReq)
    if !success {
        return types.Response{
            Type:    types.ConfigurationResponseType,
            Payload: types.ConfigurationResponse{},
        }
    }

	err := data.UpdateAgentCheckIn(configReq.AgentID)
	if err != nil {
		logger.Logf(logger.Error, "Could not update last callback for %v, %v", configReq.AgentID, err)
	}

    return types.Response{
        Type:    types.ConfigurationResponseType,
        Payload: configResponse,
    }
}


func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var request types.Request
	if err := decoder.Decode(&request); err != nil {
		logger.Logf(logger.Error, "Failed to decode request: %v", err)
		return
	}

	handler, exists := s.handlers[request.Type]
	if !exists {
		logger.Logf(logger.Warning, "Unknown request type: %v", request.Type)
		return
	}

	response := handler.Handle(request, conn)
	if err := encoder.Encode(response); err != nil {
		logger.Logf(logger.Warning, "Failed to send response: %v", err)
	}
}

func (s *Server) Start() {
	c2serverip := os.Getenv("C2SERVER_IP")
	c2serverport := os.Getenv("C2SERVER_PORT")

	cer, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		logger.Logf(logger.Error, "Failed to load certificates: %v", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	listener, err := tls.Listen("tcp", c2serverip+":"+c2serverport, config)
	if err != nil {
		logger.Logf(logger.Error, "Failed to start server: %v", err)
	}
	defer listener.Close()
	logger.Logf(logger.Info, "Started server listening on %v:%v", c2serverip, c2serverport)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Logf(logger.Error, "Error accepting connection")
			continue
		}
		go s.handleConnection(conn)
	}
}

func NewServer() *Server {
    return &Server{
        handlers: map[types.RequestType]Handler{
            types.ConfigurationRequestType: &ConfigurationHandler{},
        },
    }
}

func Init() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("logs/server.log")
	if err != nil {
		log.Fatalf("Error setting log file: %v\n", err)
	}

	data.OpenDatabase()
	data.InitDatabase()

	gob.Register(types.Request{})
    gob.Register(types.ConfigurationRequest{})
    gob.Register(types.ConfigurationResponse{})
}

type Handler interface {
	Handle(request types.Request, conn net.Conn) types.Response
}

type Server struct {
	handlers map[types.RequestType]Handler
}

func main() {
	Init()
	server := NewServer()
	server.Start()
}
