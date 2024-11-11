package client_utils

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"

	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

func HandleFileRequest(beacon *tls.Conn, encoder *gob.Encoder, decoder *gob.Decoder, agentID string) error {
	logger.Logf(logger.Info, "Sending file request")

	for {
		if err := SendRequest(encoder, types.FileRequestType, types.FileRequest{AgentID: agentID}); err != nil {
			return err
		}
		var response types.Response
		if err := decoder.Decode(&response); err != nil {
			return fmt.Errorf("error decoding command response: %v", err)
		}
		if response.Type == types.FileResponseType {
			if fileResponse, ok := response.Payload.(types.FileResponse); ok {
				logger.Logf(logger.Info, "Recieved File Response Type: FileID: %v AgentID: %v Type: %v", fileResponse.FileID, fileResponse.AgentID, fileResponse.Type)
				goto Exit	
			} else {
				return fmt.Errorf("unexpected payload type for CommandResponse")
			}			
		} else {
			return fmt.Errorf("unexpected response type: %v, expected CommandResponseType", response.Type)
		}
	}
Exit:
	return nil
}
