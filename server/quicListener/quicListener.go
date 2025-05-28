package quicListener

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	quic "github.com/quic-go/quic-go"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/server/handlers"
	"github.com/PatronC2/Patron/types"
)

func (s *Server) Start() error {
	c2serverip := os.Getenv("QUIC_LISTENER_IP")
	c2serverport := os.Getenv("QUIC_LISTENER_PORT")

	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		return fmt.Errorf("failed to load cert: %w", err)
	}

	quicConfig := &quic.Config{}
	listener, err := quic.ListenAddr(c2serverip+":"+c2serverport, &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"quic-patron"}}, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to start QUIC listener: %w", err)
	}
	logger.Logf(logger.Info, "QUIC listener started on port %v", c2serverport)

	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			logger.Logf(logger.Error, "accept error: %v", err)
			continue
		}
		go s.handleSession(sess)
	}
}

func (s *Server) handleSession(sess quic.Connection) {
	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			logger.Logf(logger.Warning, "stream accept error: %v", err)
			return
		}
		go s.handleStream(stream)
	}
}

func (s *Server) handleStream(stream quic.Stream) {
	defer stream.Close()

	request := &patronobuf.Request{}
	if err := common.ReadDelimited(stream, request); err != nil {
		logger.Logf(logger.Warning, "failed to read request: %v", err)
		return
	}

	handler, exists := s.handlers[request.Type]
	if !exists {
		logger.Logf(logger.Warning, "unknown request type: %v", request.Type)
		return
	}

	response := handler.Handle(request, stream)
	if err := common.WriteDelimited(stream, response); err != nil {
		logger.Logf(logger.Warning, "failed to write response: %v", err)
	}
}

type Handler interface {
	Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response
}

type Server struct {
	handlers map[patronobuf.RequestType]Handler
}

func NewServer() *Server {
	return &Server{
		handlers: map[patronobuf.RequestType]Handler{
			patronobuf.RequestType_CONFIGURATION: &handlers.ConfigurationHandler{},
		},
	}
}
