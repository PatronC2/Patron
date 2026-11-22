package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"google.golang.org/protobuf/proto"
)

func WriteDelimited(w io.Writer, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	fullMsg := append(lengthBuf, data...)

	_, err = w.Write(fullMsg)
	return err
}

func ReadDelimited(r io.Reader, msg proto.Message) error {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	buf := make([]byte, length)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return fmt.Errorf("read %d/%d bytes: %w", n, length, err)
	}

	return proto.Unmarshal(buf, msg)
}


func ReadDelimitedBytes(r io.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}
	msgLen := binary.BigEndian.Uint32(lenBuf[:])
	if msgLen == 0 {
		return []byte{}, nil
	}
	data := make([]byte, msgLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}
	return data, nil
}

type delimitedFrame struct {
	data []byte
	err  error
}

// ReadDelimitedBytesWithTimeout is ReadDelimitedBytes with a hard deadline. A
// blocking stream read (desynchronized or dead peer) can only be interrupted
// from the outside, so the read runs in a goroutine raced against the timeout.
// On a healthy connection the frame arrives in milliseconds; the deadline only
// fires when something is actually wrong.
func ReadDelimitedBytesWithTimeout(r io.Reader, timeout time.Duration) ([]byte, error) {
	ch := make(chan delimitedFrame, 1)
	go func() {
		data, err := ReadDelimitedBytes(r)
		ch <- delimitedFrame{data: data, err: err}
	}()
	select {
	case f := <-ch:
		return f.data, f.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out after %v waiting for message", timeout)
	}
}
