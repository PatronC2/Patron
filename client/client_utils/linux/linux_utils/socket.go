package linux_utils

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// Handler is called for each newline-delimited message received.
type Handler func(msg string) error

type Server struct {
	path    string
	handler Handler

	ln net.Listener

	wg       sync.WaitGroup
	stopOnce sync.Once
}

// New creates a server that listens on socketPath and invokes handler for each line.
func New(socketPath string, handler Handler) *Server {
	return &Server{
		path:    socketPath,
		handler: handler,
	}
}

// Start begins listening and accepting connections.
// It returns once the listener is created; accept loop runs in a goroutine.
// Call Stop() (or cancel ctx in Run) to shut down.
func (s *Server) Start() error {
	if s.handler == nil {
		return errors.New("unixsock: handler is nil")
	}

	// Remove stale socket file if present
	_ = os.Remove(s.path)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	ln, err := net.Listen("unix", s.path)
	if err != nil {
		return err
	}
	s.ln = ln

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Run is a convenience helper: Start, then block until ctx is done, then Stop.
func (s *Server) Run(ctx context.Context) error {
	if err := s.Start(); err != nil {
		return err
	}
	<-ctx.Done()
	return s.Stop()
}

// Stop closes the listener and removes the socket file.
// It waits for all connection goroutines to finish.
func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		if s.ln != nil {
			err = s.ln.Close()
		}
		_ = os.Remove(s.path)
	})
	s.wg.Wait()
	return err
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			// Listener closed -> exit loop
			return
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return // client closed or error
		}

		_ = s.handler(line) // handler decides what to do
	}
}
