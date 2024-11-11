package main

import (
	"crypto/tls"
	"encoding/gob"
	"log"
	"net"
	"os"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/server/handlers"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/lib/common"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	for {
		var request types.Request
		err := decoder.Decode(&request)
		if err != nil {
			if err.Error() == "EOF" {
				logger.Logf(logger.Info, "Client disconnected")
				return
			}
			logger.Logf(logger.Error, "Failed to decode request: %v", err)
			return
		}

		handler, exists := s.handlers[request.Type]
		if !exists {
			logger.Logf(logger.Warning, "Unknown request type: %v", request.Type)
			continue
		}

		response := handler.Handle(request, conn)
		if err := encoder.Encode(response); err != nil {
			logger.Logf(logger.Warning, "Failed to send response: %v", err)
			return
		}
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

// When new functionalities are added, add the new types and handlers to the lists in here
func NewServer() *Server {
	common.RegisterGobTypes()

    return &Server{
        handlers: map[types.RequestType]Handler{
            types.ConfigurationRequestType:		&handlers.ConfigurationHandler{},
			types.CommandRequestType:			&handlers.CommandHandler{},
			types.CommandStatusRequestType:		&handlers.CommandStatusHandler{},
			types.KeysRequestType:				&handlers.KeysHandler{},
			types.FileRequestType:				&handlers.FileRequestHandler{},
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
}

// Unique types required in the main. Do not move to types package, it won't work.
type Handler interface {
	Handle(request types.Request, conn net.Conn) types.Response
}

type Server struct {
	handlers map[types.RequestType]Handler
}

func main() {
	Init()
	data.OpenDatabase()
	data.InitDatabase()
	server := NewServer()
	server.Start()
}
