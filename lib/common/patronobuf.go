package common

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
)

func WriteDelimited(w io.Writer, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func ReadDelimited(r io.Reader, msg proto.Message) error {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	return proto.Unmarshal(buf, msg)
}
