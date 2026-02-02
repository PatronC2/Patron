package windows_utils

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"

	"github.com/Microsoft/go-winio"
)

// Handler is called for each newline-delimited message received.
type Handler func(msg string) error

type Server struct {
	name    string
	handler Handler

	ln net.Listener

	wg       sync.WaitGroup
	stopOnce sync.Once
}

// New creates a server that listens on \\.\pipe\<name> and invokes handler for each line.
// Pass either "bang_log" or full "\\\\.\\pipe\\bang_log".
func New(pipeName string, handler Handler) *Server {
	// normalize
	full := pipeName
	if len(full) < 9 || full[:9] != `\\.\pipe\` {
		full = `\\.\pipe\` + pipeName
	}
	return &Server{name: full, handler: handler}
}

// Start begins listening and accepting connections.
func (s *Server) Start() error {
	if s.handler == nil {
		return errors.New("winpipe: handler is nil")
	}

	ln, err := winio.ListenPipe(s.name, &winio.PipeConfig{
		// Restrictive by default: LocalSystem + Administrators.
		// If your client runs as a normal user and gets ACCESS_DENIED,
		// loosen this to include that user/group.
		SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;BA)",
		MessageMode:        false, // byte stream mode
		InputBufferSize:    64 * 1024,
		OutputBufferSize:   64 * 1024,
	})
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

// Stop closes the listener and waits for all goroutines to finish.
func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		if s.ln != nil {
			err = s.ln.Close()
		}
	})
	s.wg.Wait()
	return err
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			// listener closed
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
			return
		}
		_ = s.handler(line)
	}
}
