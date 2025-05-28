package tcpListener

import (
	"crypto/tls"
	"io"
	"net"
	"os"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/server/handlers"
	"github.com/PatronC2/Patron/types"
	"google.golang.org/protobuf/proto"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	read := func(msg proto.Message) error {
		return common.ReadDelimited(conn, msg)
	}
	write := func(msg proto.Message) error {
		return common.WriteDelimited(conn, msg)
	}

	for {
		request := &patronobuf.Request{}

		if err := read(request); err != nil {
			if err == io.EOF {
				logger.Logf(logger.Info, "Client disconnected")
				return
			}
			logger.Logf(logger.Error, "Failed to decode protobuf request: %v", err)
			return
		}

		logger.Logf(logger.Debug, "Received request of type: %v", request.Type)

		handler, exists := s.handlers[request.Type]
		if !exists {
			logger.Logf(logger.Warning, "Unknown request type: %v", request.Type)
			return
		}

		response := handler.Handle(request, conn)

		if err := write(response); err != nil {
			logger.Logf(logger.Warning, "Failed to send response: %v", err)
			return
		}
	}
}

func (s *Server) Start() {
	c2serverip := os.Getenv("TCP_LISTENER_IP")
	c2serverport := os.Getenv("TCP_LISTENER_PORT")

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

	return &Server{
		handlers: map[patronobuf.RequestType]Handler{
			patronobuf.RequestType_CONFIGURATION:  &handlers.ConfigurationHandler{},
			patronobuf.RequestType_COMMAND:        &handlers.CommandHandler{},
			patronobuf.RequestType_COMMAND_STATUS: &handlers.CommandStatusHandler{},
			patronobuf.RequestType_KEYS:           &handlers.KeysHandler{},
			patronobuf.RequestType_FILE:           &handlers.FileRequestHandler{},
			patronobuf.RequestType_FILE_TO_SERVER: &handlers.FileToServerHandler{},
		},
	}
}

type Handler interface {
	Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response
}

type Server struct {
	handlers map[patronobuf.RequestType]Handler
}
